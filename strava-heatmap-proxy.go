package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strconv"
	"time"
)

type Param struct {
	CookiesFile *string
	Port        *string
	Target      *string
	NoInit      *bool
	Verbose     *bool
}

func getParam() *Param {
	cookiesfile, err := os.UserHomeDir()
	if err != nil {
		cookiesfile = "cookies.json"
	} else {
		cookiesfile = path.Join(cookiesfile, ".config", "strava-heatmap-proxy", "strava-cookies.json")
	}
	param := &Param{
		CookiesFile: flag.String("cookies", cookiesfile, "Path to the cookies file"),
		Port:        flag.String("port", "8080", "Local proxy port"),
		Target:      flag.String("target", "https://content-a.strava.com/", "Heatmap provider URL"),
		NoInit:      flag.Bool("no-init", false, "Don't try to refresh CloudFront cookies on startup"),
		Verbose:     flag.Bool("verbose", false, "Verbose logging"),
	}
	flag.Parse()
	return param
}

type cookieEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type StravaSessionClient struct {
	sessionIdentifier           string
	cloudFrontCookies           []*http.Cookie
	cloudFrontCookiesExpiration time.Time
}

func NewStravaSessionClient(cookiesFilePath string) (*StravaSessionClient, error) {
	file, err := os.Open(cookiesFilePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var entries []cookieEntry
	if err := json.NewDecoder(file).Decode(&entries); err != nil {
		return nil, fmt.Errorf("error decoding json: %w", err)
	}

	client := &StravaSessionClient{}
	for _, entry := range entries {
		if entry.Name == "_strava4_session" {
			client.sessionIdentifier = entry.Value
			break
		}
	}
	if client.sessionIdentifier == "" {
		return nil, fmt.Errorf("_strava4_session not found in cookies file")
	}

	if err := client.readCloudFrontCookiesFromFile(entries); err != nil {
		log.Printf("No valid CloudFront cookies found in file: %v", err)
	}

	return client, nil
}

func (c *StravaSessionClient) readCloudFrontCookiesFromFile(entries []cookieEntry) error {
	var cookies []*http.Cookie
	var expiration int64

	for _, entry := range entries {
		switch entry.Name {
		case "CloudFront-Signature", "CloudFront-Policy", "CloudFront-Key-Pair-Id", "_strava_idcf":
			cookies = append(cookies, &http.Cookie{
				Name:  entry.Name,
				Value: entry.Value,
			})
		case "_strava_CloudFront-Expires":
			var err error
			expiration, err = strconv.ParseInt(entry.Value, 10, 64)
			if err != nil {
				log.Printf("Invalid timestamp value for %s: %s", entry.Name, entry.Value)
			}
		}
	}

	if len(cookies) < 4 {
		return fmt.Errorf("not all required CloudFront cookies found in file")
	}

	c.cloudFrontCookies = cookies
	if expiration != 0 {
		c.cloudFrontCookiesExpiration = time.UnixMilli(expiration)
		log.Printf("CloudFront cookies from file will expire at %s", c.cloudFrontCookiesExpiration)
	}

	return nil
}

func (c *StravaSessionClient) fetchCloudFrontCookies() error {
	req, err := http.NewRequest("HEAD", "https://www.strava.com/maps", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Cookie", fmt.Sprintf("_strava4_session=%s;", c.sessionIdentifier))

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to perform request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var cookies []*http.Cookie
	var expiration int64

	for _, cookie := range resp.Cookies() {
		switch cookie.Name {
		case "CloudFront-Signature", "CloudFront-Policy", "CloudFront-Key-Pair-Id", "_strava_idcf":
			cookies = append(cookies, cookie)
		case "_strava_CloudFront-Expires":
			expiration, err = strconv.ParseInt(cookie.Value, 10, 64)
			if err != nil {
				log.Printf("Invalid timestamp value for %s: %s", cookie.Name, cookie.Value)
			}
		}
	}

	if len(cookies) < 4 {
		return fmt.Errorf("not all required CloudFront cookies received")
	}

	c.cloudFrontCookies = cookies
	if expiration != 0 {
		c.cloudFrontCookiesExpiration = time.UnixMilli(expiration)
		log.Printf("New CloudFront cookies will expire at %s", c.cloudFrontCookiesExpiration)
	}

	return nil
}

func main() {
	param := getParam()
	target, err := url.Parse(*param.Target)
	if err != nil {
		log.Fatalf("Could not parse target url: %s", err)
	}

	client, err := NewStravaSessionClient(*param.CookiesFile)
	if err != nil {
		log.Fatalf("Could not initialize Strava client: %s", err)
	}

	if len(client.cloudFrontCookies) == 0 || (!*param.NoInit && time.Now().After(client.cloudFrontCookiesExpiration)) {
		log.Printf("Fetching new CloudFront cookies...")
		if err := client.fetchCloudFrontCookies(); err != nil {
			log.Fatalf("Warning: Failed to fetch CloudFront cookies: %s", err)
		}
	}

	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		// refresh expired CloudFront cookies before forwarding the request
		if client.cloudFrontCookiesExpiration.IsZero() || time.Now().After(client.cloudFrontCookiesExpiration) {
			log.Printf("CloudFront cookies have expired, refreshing...")
			if err := client.fetchCloudFrontCookies(); err != nil {
				log.Fatalf("Warning: Failed to fetch CloudFront cookies: %s", err)
			}
		}
		// add CloudFront cookies to the request
		for _, c := range client.cloudFrontCookies {
			req.AddCookie(c)
		}
		if *param.Verbose {
			log.Printf("Got request: %s", req.URL)
		}
	}

	proxy := httputil.ReverseProxy{Director: director}
	http.Handle("/", &proxy)
	log.Printf("Started proxy for target %s on http://localhost:%s/ ..", *param.Target, *param.Port)
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
