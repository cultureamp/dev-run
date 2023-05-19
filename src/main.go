package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Repositories []string `yaml:"repositories"`
	Token        string   `yaml:"token"`
}

type ServiceInfo struct {
	Repo    string
	Service string
}

type DockerCompose struct {
	Services map[string]interface{} `yaml:"services"`
}

const (
	commonNetwork = "common_network"
	downloadDir   = "../downloads"
	configFile    = "../repos.yaml"
)

func main() {
	action := flag.String("action", "clone", "Action to perform: 'clone', 'docker-up', 'list-services', or 'run-service'")
	targetService := flag.String("service", "", "Target service to run")
	flag.Parse()

	config, err := loadConfig(configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v\n", err)
	}

	switch *action {
	case "clone":
		cloneRepositories(config, downloadDir)
	case "docker-up":
		runDockerCompose(config, downloadDir)
	case "list-services":
		services, err := listServices(config, downloadDir)
		if err != nil {
			log.Fatalf("Failed to list services: %v\n", err)
		}
		fmt.Printf("Services:\n")
		for _, service := range services {
			fmt.Printf("- Repo: %s, Service: %s\n", service.Repo, service.Service)
		}
	case "run-service":
		if *targetService == "" {
			log.Fatalf("Please provide a service name using the -service flag\n")
		}
		err = runTargetService(config, downloadDir, *targetService)
		if err != nil {
			log.Fatalf("Failed to run target service: %v\n", err)
		}
	default:
		log.Fatalf("Invalid action: %s. Choose 'clone', 'docker-up', 'list-services', or 'run-service'.\n", *action)
	}
}

func loadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
