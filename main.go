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

func NewStravaLoginRequest() (*http.Request, error) {
	return http.NewRequest("GET", "https://www.strava.com/login", nil)
}

func NewStravaSessionRequest(email string, password string, token string) (*http.Request, error) {
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
	return req, err
}

func NewStravaAuthRequest() (*http.Request, error) {
	return http.NewRequest("GET", "https://heatmap-external-a.strava.com/auth", nil)
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

type StravaClient struct {
	http.Client
}

func NewStravaClient() *StravaClient {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Failed to create cookie jar: %s", err)
	}
	return &StravaClient{http.Client{Jar: jar}}
}

func (client *StravaClient) Send(req *http.Request, err error) (*http.Response, error) {
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Request '%s' failed with status code %s", req.RequestURI, resp.Status)
	}
	return resp, nil
}

func (client *StravaClient) Authenticate(email string, password string) error {
	res, err := client.Send(NewStravaLoginRequest())
	if err != nil {
		return fmt.Errorf("Could not send login request: %w", err)
	}
	token, err := ExtractAuthenticityToken(res)
	if err != nil {
		return fmt.Errorf("Could not get authenticity token for login form: %w", err)
	}
	if _, err = client.Send(NewStravaSessionRequest(email, password, token)); err != nil {
		return fmt.Errorf("Could not send session request: %w", err)
	}
	if _, err = client.Send(NewStravaAuthRequest()); err != nil {
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
