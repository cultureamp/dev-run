package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
)

func runTargetService(config *Config, downloadDir, targetService string) error {
	services, err := listServices(config, downloadDir)
	if err != nil {
		return err
	}

	for _, service := range services {
		if service.Service == targetService {
			err = runServiceInRepo(service.Repo, downloadDir, targetService)
			if err != nil {
				log.Printf("Failed to run service '%s' in repository %s: %v\n", targetService, service.Repo, err)
			}
		}
	}

	return nil
}

func runServiceInRepo(repo, downloadDir, targetService string) error {
	repoName := filepath.Base(repo)
	repoName = repoName[:len(repoName)-4] // Remove .git extension
	repoPath := filepath.Join(downloadDir, repoName)

	fmt.Printf("Running service '%s' in repository: %s\n", targetService, repo)

	cmd := exec.Command("docker-compose", "up", "-d", targetService)
	cmd.Dir = repoPath
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("run service error: %v, output: %s", err, string(output))
	}

	fmt.Printf("Successfully ran service '%s' in repository: %s\n", targetService, repo)
	return nil
}
