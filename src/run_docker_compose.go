package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

func runDockerCompose(config *Config, downloadDir string) {
	err := createCommonNetwork()
	if err != nil {
		log.Fatalf("Failed to create common network: %v\n", err)
	}

	allServices, err := listServices(config, downloadDir)
	if err != nil {
		log.Fatalf("Failed to list services: %v\n", err)
	}

	var wg sync.WaitGroup
	for _, repo := range config.Repositories {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			err := updateAndRunDockerCompose(repo, downloadDir, allServices)
			if err != nil {
				log.Printf("Failed to run docker-compose in repository %s: %v\n", repo, err)
			}
		}(repo)
	}

	wg.Wait()
}

func createCommonNetwork() error {
	cmd := exec.Command("docker", "network", "create", commonNetwork)
	output, err := cmd.CombinedOutput()
	if err != nil {
		if strings.Contains(string(output), "already exists") {
			log.Println("Common network already exists")
			return nil
		}
		return fmt.Errorf("create common network error: %v, output: %s", err, string(output))
	}
	fmt.Printf("Successfully created common network: %s\n", commonNetwork)
	return nil
}

func updateAndRunDockerCompose(repo, downloadDir string, allServices []ServiceInfo) error {
	repoName := filepath.Base(repo)
	repoName = repoName[:len(repoName)-4] // Remove .git extension
	repoPath := filepath.Join(downloadDir, repoName)

	composeFile := filepath.Join(repoPath, "docker-compose.yml")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return fmt.Errorf("docker-compose.yml not found in repository: %s", repo)
	}
	data, err := os.ReadFile(composeFile)
	if err != nil {
		return err
	}

	var rawConfig map[string]interface{}
	err = yaml.Unmarshal(data, &rawConfig)
	if err != nil {
		return err
	}

	// Namespace the service names
	services := rawConfig["services"]
	if services == nil {
		return fmt.Errorf("no services found in docker-compose.yml")
	}

	servicesMap, ok := services.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid services configuration in docker-compose.yml")
	}

	namespacedServicesMap := make(map[string]interface{})
	for serviceName, serviceData := range servicesMap {
		namespacedServiceName := repoName + "_" + serviceName

		serviceMap, ok := serviceData.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid service configuration in docker-compose.yml")
		}

		// Update the depends_on field with namespaced service names
		if dependsOn, ok := serviceMap["depends_on"].([]interface{}); ok {
			namespacedDependsOn := make([]interface{}, len(dependsOn))
			for i, dep := range dependsOn {
				depName, ok := dep.(string)
				if !ok {
					return fmt.Errorf("invalid depends_on configuration in docker-compose.yml")
				}
				namespacedDependsOn[i] = repoName + "_" + depName
			}
			serviceMap["depends_on"] = namespacedDependsOn
		}

		// Add environment variables for other service URLs
		envs := make([]string, 0)
		if existingEnvs, ok := serviceMap["environment"].([]interface{}); ok {
			for _, env := range existingEnvs {
				envs = append(envs, env.(string))
			}
		}

		for _, otherService := range allServices {
			if otherService.Repo != repo {
				envVarName := strings.ToUpper(strings.ReplaceAll(otherService.Service, "-", "_")) + "_URL"
				envVarValue := fmt.Sprintf("%s=http://%s:port", envVarName, otherService.Service)
				envs = append(envs, envVarValue)
			}
		}

		serviceMap["environment"] = envs
		namespacedServicesMap[namespacedServiceName] = serviceMap

	}
	rawConfig["services"] = namespacedServicesMap

	// Add common network to the docker-compose configuration
	networks := rawConfig["networks"]
	if networks == nil {
		networks = make(map[string]interface{})
		rawConfig["networks"] = networks
	}

	networksMap, ok := networks.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid networks configuration in docker-compose.yml")
	}

	if _, exists := networksMap[commonNetwork]; !exists {
		networksMap[commonNetwork] = nil
	}

	// Add common network to each service
	services = rawConfig["services"]
	if services == nil {
		return fmt.Errorf("no services found in docker-compose.yml")
	}

	servicesMap, ok = services.(map[string]interface{})
	if !ok {
		return fmt.Errorf("invalid services configuration in docker-compose.yml")
	}

	for _, serviceData := range servicesMap {
		serviceMap, ok := serviceData.(map[string]interface{})
		if !ok {
			return fmt.Errorf("invalid service configuration in docker-compose.yml")
		}

		serviceNetworks := serviceMap["networks"]
		if serviceNetworks == nil {
			serviceNetworks = make([]interface{}, 0)
			serviceMap["networks"] = serviceNetworks
		}

		serviceNetworksList, ok := serviceNetworks.([]interface{})
		if !ok {
			return fmt.Errorf("invalid service networks configuration in docker-compose.yml")
		}

		serviceNetworksList = append(serviceNetworksList, commonNetwork)
		serviceMap["networks"] = serviceNetworksList
	}

	// Save the updated docker-compose.yml file
	updatedData, err := yaml.Marshal(rawConfig)
	if err != nil {
		return err
	}

	err = os.WriteFile(composeFile, updatedData, 0644)
	if err != nil {
		return err
	}

	cmd := exec.Command("docker-compose", "up", "-d")
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("run docker-compose error: %v, output: %s", err, string(output))
	}

	fmt.Printf("Successfully ran docker-compose in repository: %s\n", repo)
	return nil
}
