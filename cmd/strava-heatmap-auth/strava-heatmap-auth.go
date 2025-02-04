package main

import (
	"flag"
	"log"
	"net/url"
	"os"
	"path"

	"github.com/patrickziegler/strava-heatmap-proxy/internal/clients"
	"github.com/patrickziegler/strava-heatmap-proxy/internal/config"
	"github.com/patrickziegler/strava-heatmap-proxy/internal/pipe"
)

type Param struct {
	Config *string
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
		Config: flag.String("config", configfile, "Path to configuration file"),
		Target: flag.String("target", "https://heatmap-external-a.strava.com/", "Heatmap provider URL"),
	}
	flag.Parse()
	return param
}

func main() {
	param := getParam()
	cfg, err := config.ParseConfig(*param.Config)
	if err != nil {
		log.Fatalf("Failed to get configuration: %s", err)
	}
	target, err := url.Parse(*param.Target)
	if err != nil {
		log.Fatalf("Could not parse target url: %s", err)
	}
	strava := clients.NewStravaClient(target)
	if err := strava.Authenticate(cfg.Email, cfg.Password); err != nil {
		log.Fatalf("Failed to authenticate client: %s", err)
	}
	pipe.ReplaceCloudFrontTokens(strava)
}
