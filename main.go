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
	req *http.Request
	err error
}

func NewStravaClient() *StravaClient {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to create cookie jar: %s", err)
	}
	return &StravaClient{http.Client{Jar: jar}, nil, nil}
}

func (client *StravaClient) CreateStravaLoginRequest() *StravaClient {
	client.req, client.err = http.NewRequest("GET", "https://www.strava.com/login", nil)
	return client
}

func (client *StravaClient) CreateStravaSessionRequest(email string, password string, token string) *StravaClient {
	data := url.Values{}
	data.Add("utf8", "\u2713")
	data.Add("authenticity_token", token)
	data.Add("plan", "")
	data.Add("email", email)
	data.Add("password", password)
	data.Add("remember_me", "on")
	req, err := http.NewRequest("POST", "https://www.strava.com/session", strings.NewReader(data.Encode()))
	if err != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	client.req = req
	client.err = err
	return client
}

func (client *StravaClient) CreateStravaAuthRequest() *StravaClient {
	client.req, client.err = http.NewRequest("GET", "https://heatmap-external-a.strava.com/auth", nil)
	return client
}

func (client *StravaClient) Send() (*http.Response, error) {
	if client.req == nil {
		return nil, errors.New("No request available to send")
	}
	if client.err != nil {
		return nil, client.err
	}
	resp, err := client.Do(client.req)
	if err != nil {
		return nil, &ResponseError{err, resp}
	}
	client.req = nil
	client.err = nil
	return resp, nil
}

func ExtractAuthenticityToken(resp *http.Response) (string, error) {
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
	res, err := client.CreateStravaLoginRequest().Send()
	if err != nil {
		return fmt.Errorf("Could not send login request: %w", err)
	}
	token, err := ExtractAuthenticityToken(res)
	if err != nil {
		return fmt.Errorf("Could not get authenticity token for login form: %w", err)
	}
	if _, err = client.CreateStravaSessionRequest(email, password, token).Send(); err != nil {
		return fmt.Errorf("Could not send session request: %w", err)
	}
	if _, err = client.CreateStravaAuthRequest().Send(); err != nil {
		return fmt.Errorf("Could not send auth request: %w", err)
	}
	return nil
}

func (client *StravaClient) GetCloudFrontCookies() map[string]string {
	cookies := map[string]string{}
	target, _ := url.Parse("https://www.strava.com")
	for _, cookie := range client.Jar.Cookies(target) {
		if strings.HasPrefix(cookie.Name, "CloudFront") {
			cookies[cookie.Name] = cookie.Value
		}
	}
	return cookies
}

func (client *StravaClient) AddCookies(req *http.Request) {
	for _, c := range client.Jar.Cookies(req.URL) {
		req.AddCookie(c)
	}
}

type CookieClient interface {
	AddCookies(*http.Request)
}

type StravaProxy struct {
	httputil.ReverseProxy
	Client CookieClient
}

func NewStravaProxy(client CookieClient) *StravaProxy {
	target, _ := url.Parse("https://heatmap-external-a.strava.com/")
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		client.AddCookies(req)
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
		log.Fatalf("Failed to authenticate client: %s", err)
	}

	for k, v := range client.GetCloudFrontCookies() {
		fmt.Printf("%s\t%s\n", k, v)
	}

	http.Handle("/", NewStravaProxy(client))
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
