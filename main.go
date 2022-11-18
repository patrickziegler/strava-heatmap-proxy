package main

// https://github.com/erik/strava-heatmap-proxy/blob/main/scripts/refresh_strava_credentials.ts

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"
	"syscall"

	"golang.org/x/term"
)

func getUser() string {
	fmt.Print("Strava user: ")
	var user string
	_, err := fmt.Scanln(&user)
	if err != nil {
		fmt.Println("sackradding", err)
	}
	return user
}

func getPassword() string {
	fmt.Print("Strava password: ")
	bytepw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Print("\n")
	if err != nil {
		log.Fatal(err)
	}
	return string(bytepw)
}

func main() {
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
	params.Add("email", getUser())
	params.Add("password", getPassword())
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
}
