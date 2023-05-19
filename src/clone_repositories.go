package main

import (
	"fmt"
	"log"
	"os/exec"
	"sync"
)

func cloneRepositories(config *Config, downloadDir string) {
	var wg sync.WaitGroup

	for _, repo := range config.Repositories {
		wg.Add(1)
		go func(repo string) {
			defer wg.Done()
			err := cloneRepo(config.Token, repo, downloadDir)
			if err != nil {
				log.Printf("Failed to clone repository %s: %v\n", repo, err)
			}
		}(repo)
	}

	wg.Wait()
}

func cloneRepo(repo, downloadDir, token string) error {
	fmt.Printf("Cloning repository: %s\n", repo)

	cmd := exec.Command("git", "clone", fmt.Sprintf("https://%s:x-oauth-basic@github.com/%s", token, repo))
	cmd.Dir = downloadDir
	output, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("clone error: %v, output: %s", err, string(output))
	}

	fmt.Printf("Repository cloned: %s\n", repo)
	return nil
}
