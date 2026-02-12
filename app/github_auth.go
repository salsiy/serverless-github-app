package main

import (
	"fmt"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation/v2"
	"github.com/google/go-github/v57/github"
)

var (
	defaultOwner string = "salsiy"
)

func createGitHubClient(installationID int64) (*github.Client, error) {
	if githubAppPrivateKeyPem == "" {
		return nil, fmt.Errorf("GitHub App private key not loaded")
	}

	if githubAppID == 0 {
		return nil, fmt.Errorf("GitHub App ID not configured")
	}

	installationTransport, err := ghinstallation.New(
		http.DefaultTransport,
		githubAppID,
		installationID,
		[]byte(githubAppPrivateKeyPem),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create installation transport: %w", err)
	}

	client := github.NewClient(&http.Client{Transport: installationTransport})
	logger.Info("created GitHub client for installation")
	return client, nil
}
