package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"

	"github.com/patrickziegler/strava-heatmap-proxy/internal/strava"
)

type Param struct {
	Config *string
	Port   *string
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
		Port:   flag.String("port", "8080", "Local proxy port"),
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
	client := strava.NewStravaClient()
	if err = client.Authenticate(config.Email, config.Password); err != nil {
		log.Fatalf("Failed to authenticate client: %s", err)
	}
	for k, v := range client.GetCloudFrontCookies() {
		fmt.Printf("%s\t%s\n", k, v)
	}
	log.Printf("Starting strava heatmap proxy on port %s ..", *param.Port)
	http.Handle("/", strava.NewStravaProxy(client))
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
