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
	}
	flag.Parse()
	return param
}

type cookieEntry struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func readCookies(path string) ([]*http.Cookie, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	var entries []cookieEntry
	if err := json.NewDecoder(file).Decode(&entries); err != nil {
		return nil, fmt.Errorf("error decoding json: %w", err)
	}

	targetNames := map[string]bool{
		"CloudFront-Signature":   true,
		"CloudFront-Policy":      true,
		"CloudFront-Key-Pair-Id": true,
		"_strava_idcf":           true,
	}

	var expiry time.Time
	var cookies []*http.Cookie
	for _, entry := range entries {
		if targetNames[entry.Name] {
			cookie := &http.Cookie{
				Name:  entry.Name,
				Value: entry.Value,
			}
			cookies = append(cookies, cookie)
			log.Printf("%s\t%s", entry.Name, entry.Value)
		} else if entry.Name == "_strava_CloudFront-Expires" {
			timestamp, err := strconv.ParseInt(entry.Value, 10, 64) // parse unix timestamp in milliseconds
			if err != nil {
				log.Printf("Invalid timestamp value for %s: %s", entry.Name, entry.Value)
			} else {
				expiry = time.UnixMilli(timestamp)
			}
		}
	}

	if !expiry.IsZero() {
		if expiry.Before(time.Now()) {
			log.Fatalf("Cookies have expired %s)", expiry)
		} else {
			log.Printf("Cookies will expire %s)", expiry)
		}
	}

	return cookies, nil
}

func main() {
	param := getParam()
	target, err := url.Parse(*param.Target)
	if err != nil {
		log.Fatalf("Could not parse target url: %s", err)
	}
	cookies, err := readCookies(*param.CookiesFile)
	if err != nil {
		log.Fatalf("Could not parse target url: %s", err)
	}
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		req.Host = target.Host
		for _, c := range cookies {
			req.AddCookie(c)
		}
		// log.Printf("Got request: %s", req.URL)
	}
	proxy := httputil.ReverseProxy{Director: director}
	http.Handle("/", &proxy)
	log.Printf("Starting proxy for target %s on http://localhost:%s/ ..", *param.Target, *param.Port)
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
