package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/patrickziegler/strava-heatmap-proxy/internal/strava"
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
	config, err := strava.ParseConfig(*param.Config)
	if err != nil {
		log.Fatalf("Failed to get configuration: %s", err)
	}
	target, err := url.Parse(*param.Target)
	if err != nil {
		log.Fatalf("Could not parse target url: %s", err)
	}
	client := strava.NewStravaClient(target)
	if err := client.Authenticate(config.Email, config.Password); err != nil {
		log.Fatalf("Failed to authenticate client: %s", err)
	}
	cookies := client.GetCloudFrontCookies()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.Replace(line, "%CloudFront-Key-Pair-Id%", cookies["CloudFront-Key-Pair-Id"], -1)
		line = strings.Replace(line, "%CloudFront-Policy%", cookies["CloudFront-Policy"], -1)
		line = strings.Replace(line, "%CloudFront-Signature%", cookies["CloudFront-Signature"], -1)
		fmt.Println(line)
	}
}
