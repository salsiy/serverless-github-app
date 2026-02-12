package main

import (
	"context"
	"fmt"

	"go.uber.org/zap"
)

var (
	// supportedEvents defines which webhook events are handled
	supportedEvents = map[string]bool{
		"release": true,
	}
)

// processWebhook processes the webhook and sends repository dispatches based on config
func processWebhook(ctx context.Context, payload *WebhookPayload) error {

	client, err := createGitHubClient(payload.Installation.ID)
	if err != nil {
		return fmt.Errorf("failed to create GitHub client: %w", err)
	}

	// Load dispatch configuration from the source repository
	config, err := loadAppConfig(ctx, client, payload.Repository.Owner.Login, payload.Repository.Name)
	if err != nil {
		return fmt.Errorf("failed to load app config: %w", err)
	}


	eventType, _ := determineEventType(payload)

	logger.Info("processing webhook event",
		zap.String("eventType", eventType),
		zap.String("action", payload.Action),
		zap.String("repo", payload.Repository.FullName),
	)

	// Find matching dispatch rules
	var dispatched int
	for _, rule := range config.Dispatches {
		if !matchesRule(rule, eventType) {
			continue
		}

		// Send dispatches to all targets
		for _, target := range rule.Targets {
			if err := sendRepositoryDispatch(ctx, client, target, payload); err != nil {
				logger.Error("failed to send repository dispatch",
					zap.Error(err),
					zap.String("target", fmt.Sprintf("%s/%s", payload.Repository.Owner.Login, target.Repo)),
				)
				continue
			}
			dispatched++
		}
	}

	logger.Info("webhook processing complete",
		zap.Int("dispatchesSent", dispatched),
	)

	return nil
}

// determineEventType determines the event type from the payload
func determineEventType(payload *WebhookPayload) (string, error) {
	if payload.Release != nil {
		eventType := "release"
		if supportedEvents[eventType] {
			return eventType, nil
		}
	}
	return "", fmt.Errorf("unsupported or unknown event type")
}

// matchesRule checks if the webhook matches the rule
func matchesRule(rule Rule, eventType string) bool {
	return rule.Event == eventType
}
