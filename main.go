package main

import (
	"encoding/json"
	"errors"
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

type ResponseError struct {
	err  error
	resp *http.Response
}

func (err *ResponseError) Error() string {
	if err.resp.StatusCode != http.StatusOK {
		return "Request " + err.resp.Request.URL.RawQuery + " returned: " + err.resp.Status
	} else {
		return "Request failed: " + err.err.Error()
	}
}

type StravaClient struct {
	http.Client
}

func NewStravaClient() *StravaClient {
	jar, _ := cookiejar.New(nil)
	return &StravaClient{http.Client{Jar: jar}}
}

func extractAuthenticityToken(resp *http.Response) (string, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("Failed to read body: %w", err)
	}
	expr, _ := regexp.Compile("name=\"authenticity_token\".*value=\"(.+?)\"")
	matches := expr.FindStringSubmatch(string(body))
	if len(matches) < 2 {
		return "", errors.New("Token not found")
	}
	return matches[1], nil
}

func (client *StravaClient) Authenticate(email string, password string) error {
	resp, err := client.Get("https://www.strava.com/login")
	if err != nil {
		return &ResponseError{err, resp}
	}
	token, err := extractAuthenticityToken(resp)
	if err != nil {
		return fmt.Errorf("Could not get authenticity token for login form: %w", err)
	}
	params := url.Values{}
	params.Add("utf8", "\u2713")
	params.Add("authenticity_token", token)
	params.Add("plan", "")
	params.Add("email", email)
	params.Add("password", password)
	params.Add("remember_me", "on")
	if resp, err := client.PostForm("https://www.strava.com/session", params); err != nil {
		return &ResponseError{err, resp}
	}
	if resp, err := client.Get("https://heatmap-external-a.strava.com/auth"); err != nil {
		return &ResponseError{err, resp}
	}
	return nil
}

func (client *StravaClient) AddAuthenticationToken(req *http.Request) {
	for _, c := range client.Jar.Cookies(req.URL) {
		req.AddCookie(c)
	}
}

func (client *StravaClient) PrintAuthenticationToken() {
	target, _ := url.Parse("https://www.strava.com")
	for _, cookie := range client.Jar.Cookies(target) {
		if strings.HasPrefix(cookie.Name, "CloudFront") {
			fmt.Println(cookie.Name, "\t", cookie.Value)
		}
	}
}

type AuthenticationClient interface {
	AddAuthenticationToken(*http.Request)
}

type StravaProxy struct {
	httputil.ReverseProxy
	Client AuthenticationClient
}

func NewStravaProxy(client AuthenticationClient) *StravaProxy {
	target, _ := url.Parse("https://heatmap-external-a.strava.com/")
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		client.AddAuthenticationToken(req)
	}
	return &StravaProxy{
		httputil.ReverseProxy{Director: director},
		client,
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

	client := NewStravaClient()
	err = client.Authenticate(config.Email, config.Password)
	if err != nil {
		panic("Failed to authenticate: " + err.Error())
	}
	client.PrintAuthenticationToken()

	proxy := NewStravaProxy(client)

	http.Handle("/", proxy)

	log.Printf("Starting proxy")
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
