package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"path"
	"strings"

	"github.com/patrickziegler/strava-heatmap-proxy/internal/strava"
)

func getConfig() *strava.Config {
	configfile, err := os.UserHomeDir()
	if err != nil {
		configfile = "config.json"
	} else {
		configfile = path.Join(configfile, ".config", "strava-heatmap-proxy", "config.json")
	}
	flag.StringVar(&configfile, "config", configfile, "Path to configuration file")
	flag.Parse()
	config, err := strava.ParseConfig(configfile)
	if err != nil {
		log.Fatalf("Failed to get configuration: %s", err)
	}
	return config
}

func main() {
	config := getConfig()
	client := strava.NewStravaClient()
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
