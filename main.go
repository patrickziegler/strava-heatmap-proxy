package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
)

type StravaProxy struct {
	httputil.ReverseProxy
	Client *http.Client
}

func NewStravaProxy() *StravaProxy {
	jar, err := cookiejar.New(nil)
	if err != nil {
		// error handling
	}

	client := &http.Client{
		Jar: jar,
	}

	target, err := url.Parse("https://heatmap-external-a.strava.com/")
	if err != nil {
		log.Fatal(err)
	}

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		for _, c := range jar.Cookies(req.URL) {
			req.AddCookie(c)
		}
	}

	return &StravaProxy{
		httputil.ReverseProxy{Director: director},
		client,
	}
}

func (proxy *StravaProxy) Authenticate(email string, password string) {
	resp, err := proxy.Client.Get("https://www.strava.com/login")
	if err != nil {
		print(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}
	text := string(body)
	exp, err := regexp.Compile("name=\"authenticity_token\".*value=\"(.+?)\"")
	authenticity_token := exp.FindStringSubmatch(text)[1]

	params := url.Values{}
	params.Add("utf8", "\u2713")
	params.Add("authenticity_token", authenticity_token)
	params.Add("plan", "")
	params.Add("email", email)
	params.Add("password", password)
	params.Add("remember_me", "on")

	resp, err = proxy.Client.PostForm("https://www.strava.com/session", params)
	if err != nil {
		fmt.Println("leck mi")
		print(err)
	} else if resp.StatusCode != http.StatusOK {
		fmt.Println("sackrement")
	}

	resp, err = proxy.Client.Get("https://heatmap-external-a.strava.com/auth")
	if err != nil {
		fmt.Println("sackrement")
		print(err)
	} else if resp.StatusCode != http.StatusOK {
		fmt.Println("zefix")
	}
}

type Config struct {
	Email    string
	Password string
}

type Param struct {
	Config *string
	Port   *string
}

func main() {
	param := &Param{
		Config: flag.String("config", "config.json", "Path to configuration file"),
		Port:   flag.String("port", "8080", "Local proxy port"),
	}
	flag.Parse()
	configFile, _ := os.Open(*param.Config)
	byt, err := io.ReadAll(configFile)
	if err != nil {
		fmt.Println("Failed to read config file: ", err)
		os.Exit(1)
	}
	var config Config
	err = json.Unmarshal(byt, &config)
	if err != nil {
		fmt.Println("Failed to parse config file: ", err)
		os.Exit(1)
	}
	if config.Email == "" {
		fmt.Println("Cannot find 'Email' in config")
		os.Exit(1)
	}
	if config.Password == "" {
		fmt.Println("Cannot find 'Password' in config")
		os.Exit(1)
	}

	proxy := NewStravaProxy()
	proxy.Authenticate(config.Email, config.Password)

	http.Handle("/", proxy)

	log.Printf("Starting proxy")
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
