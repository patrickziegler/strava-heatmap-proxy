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

func ParseConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("Could not open config file '%s': %w", path, err)
	}
	body, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Could not read from config file '%s': %w", path, err)
	}
	var config Config
	err = json.Unmarshal(body, &config)
	if err != nil {
		return nil, fmt.Errorf("Could not unmarshal json from config file '%s': %w", path, err)
	}
	if config.Email == "" {
		return nil, fmt.Errorf("Mandatory field 'Email' not found in '%s'", path)
	}
	if config.Password == "" {
		return nil, fmt.Errorf("Mandatory field 'Password' not found in '%s'", path)
	}
	return &config, nil
}

func main() {
	param := &Param{
		Config: flag.String("config", "config.json", "Path to configuration file"),
		Port:   flag.String("port", "8080", "Local proxy port"),
	}
	flag.Parse()

	config, err := ParseConfig(*param.Config)
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
	log.Printf("Starting heatmap proxy on port %s ..", *param.Port)

	http.Handle("/", strava.NewStravaProxy(client))
	log.Fatal(http.ListenAndServe(":"+*param.Port, nil))
}
