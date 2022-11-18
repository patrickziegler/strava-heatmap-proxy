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
	"strings"
)

func Doit(email string, password string, port string) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		// error handling
	}

	client := &http.Client{
		Jar: jar,
	}

	resp, err := client.Get("https://www.strava.com/login")
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

	resp2, err := client.PostForm("https://www.strava.com/session", params)
	if err != nil {
		fmt.Println("leck mi")
		print(err)
	} else if resp2.StatusCode != http.StatusOK {
		fmt.Println("sackrement")
	}

	resp3, err := client.Get("https://heatmap-external-a.strava.com/auth")
	if err != nil {
		fmt.Println("sackrement")
		print(err)
	} else if resp3.StatusCode != http.StatusOK {
		fmt.Println("zefix")
	}

	urlObj, _ := url.Parse("https://www.strava.com")
	for _, cookie := range jar.Cookies(urlObj) {
		if strings.HasPrefix(cookie.Name, "CloudFront") {
			fmt.Println(cookie.Name, ":\t", cookie.Value)
		}
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

	proxy := &httputil.ReverseProxy{
		Director: director,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		proxy.ServeHTTP(w, r)
		log.Printf("%s %s", r.Method, r.URL.Path)
	})

	log.Printf("Starting proxy")
	log.Fatal(http.ListenAndServe(":"+port, nil))
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
	Doit(config.Email, config.Password, *param.Port)
}
