package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"

	"github.com/patrickziegler/strava-heatmap-proxy/internal/clients"
	"github.com/patrickziegler/strava-heatmap-proxy/internal/config"
	"github.com/patrickziegler/strava-heatmap-proxy/internal/proxy"
)

type Param struct {
	Client *string
	Config *string
	Port   *string
	Target *string
}

func getParam() *Param {
	configfile, err := os.UserHomeDir()
	if err != nil {
		configfile = "config.json"
	} else {
		configfile = path.Join(configfile, ".config", "strava-heatmap-proxy", "config.json")
	}
	param := &Param{
		Client: flag.String("client", "strava", "Client to be used for getting CloudFront tokens, should be one of: strava, firefox"),
		Config: flag.String("config", configfile, "Path to the config file"),
		Port:   flag.String("port", "8080", "Local proxy port"),
		Target: flag.String("target", "https://heatmap-external-a.strava.com/", "Heatmap provider URL"),
	}
	flag.Parse()
	return param
}

func getFirefoxClient(target *url.URL) *clients.FirefoxClient {
	client, err := clients.NewFirefoxClient(target)
	if err != nil {
		log.Fatal("Error: ", err)
	}
	if client == nil {
		log.Fatal("Could not find CloudFront tokens in Firefox cookies")
	}
	return client
}

func getStravaClient(target *url.URL, configPath string) *clients.StravaClient {
	cfg, err := config.ParseConfig(configPath)
	if err != nil {
		log.Fatalf("Failed to get configuration: %s", err)
	}
	client := clients.NewStravaClient(target)
	if err := client.Authenticate(cfg.Email, cfg.Password); err != nil {
		log.Fatalf("Failed to authenticate client: %s", err)
	}
	return client
}

type CookieClient interface {
	proxy.CookieClient
	GetCloudFrontTokens() map[string]string
}

func main() {
	var client CookieClient
	param := getParam()
	target, err := url.Parse(*param.Target)
	if err != nil {
		log.Fatalf("Could not parse target url: %s", err)
	}
	switch *param.Client {
	case "firefox":
		client = getFirefoxClient(target)
	default:
		client = getStravaClient(target, *param.Config)
	}
	for k, v := range client.GetCloudFrontTokens() {
		fmt.Printf("%s\t%s\n", k, v)
	}
	log.Printf("Starting strava heatmap proxy on port %s ..", *param.Port)
	http.Handle("/", proxy.NewReverseProxy(client))
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
