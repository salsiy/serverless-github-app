package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v57/github"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

const (
	configFilePath = ".github/app-config.yaml"
)

func loadAppConfig(ctx context.Context, client *github.Client, owner, repo string) (*AppConfig, error) {
	logger.Info("loading app config",
		zap.String("owner", owner),
		zap.String("repo", repo),
		zap.String("path", configFilePath),
	)

	// Get the config file from the repository
	fileContent, _, _, err := client.Repositories.GetContents(
		ctx,
		owner,
		repo,
		configFilePath,
		&github.RepositoryContentGetOptions{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get config file: %w", err)
	}

	if fileContent == nil {
		return nil, fmt.Errorf("config file not found at %s", configFilePath)
	}

	// Decode the file content
	content, err := fileContent.GetContent()
	if err != nil {
		return nil, fmt.Errorf("failed to decode config file: %w", err)
	}

	// Parse with Viper
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(content)); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	var config AppConfig
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	logger.Info("app config loaded successfully",
		zap.Int("dispatches_count", len(config.Dispatches)),
	)

	// Debug log each dispatch rule
	for i, rule := range config.Dispatches {
		logger.Info("dispatch rule",
			zap.Int("rule_index", i),
			zap.String("event", rule.Event),
			zap.Int("targets_count", len(rule.Targets)),
		)
		for j, target := range rule.Targets {
			logger.Info("target",
				zap.Int("rule_index", i),
				zap.Int("target_index", j),
				zap.String("repo", target.Repo),
				zap.String("event_type", target.EventType),
			)
		}
	}

	return &config, nil
}
