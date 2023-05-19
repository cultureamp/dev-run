package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"gopkg.in/yaml.v2"
)

func listServices(config *Config, downloadDir string) ([]ServiceInfo, error) {
	var wg sync.WaitGroup
	servicesCh := make(chan ServiceInfo, len(config.Repositories))
	errCh := make(chan error, len(config.Repositories))

	for _, repo := range config.Repositories {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()

			repoServices, err := listRepoServices(repo, downloadDir)
			if err != nil {
				errCh <- err
				return
			}

			for _, service := range repoServices {
				servicesCh <- service
			}
		}(repo)
	}

	wg.Wait()
	close(servicesCh)
	close(errCh)

	var services []ServiceInfo
	for service := range servicesCh {
		services = append(services, service)
	}

	if len(errCh) > 0 {
		var errMsgs []string
		for err := range errCh {
			errMsgs = append(errMsgs, err.Error())
		}
		return services, fmt.Errorf("errors encountered: %s", strings.Join(errMsgs, "; "))
	}

	return services, nil
}

func listRepoServices(repo, downloadDir string) ([]ServiceInfo, error) {
	repoName := filepath.Base(repo)
	repoName = repoName[:len(repoName)-4] // Remove .git extension
	repoPath := filepath.Join(downloadDir, repoName)

	composeFile := filepath.Join(repoPath, "docker-compose.yml")
	if _, err := os.Stat(composeFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("docker-compose.yml not found in repository: %s", repo)
	}

	content, err := os.ReadFile(composeFile)
	if err != nil {
		return nil, err
	}

	var compose DockerCompose
	err = yaml.Unmarshal(content, &compose)
	if err != nil {
		return nil, fmt.Errorf("failed to parse YAML: %v", err)
	}

	var services []ServiceInfo
	for serviceName := range compose.Services {
		services = append(services, ServiceInfo{
			Repo:    repo,
			Service: serviceName,
		})
	}

	return services, nil
}
