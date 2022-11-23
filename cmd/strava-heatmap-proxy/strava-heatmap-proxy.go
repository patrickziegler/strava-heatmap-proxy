package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/patrickziegler/strava-heatmap-proxy/internal/strava"
)

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

	client := strava.NewStravaClient()

	err = client.Authenticate(config.Email, config.Password)
	if err != nil {
		log.Fatalf("Failed to authenticate client: %s", err)
	}

	for k, v := range client.GetCloudFrontCookies() {
		fmt.Printf("%s\t%s\n", k, v)
	}

	http.Handle("/", strava.NewStravaProxy(client))
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
