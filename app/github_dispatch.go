package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/go-github/v57/github"
	"go.uber.org/zap"
)

// sendRepositoryDispatch sends a repository dispatch event to the target repository
func sendRepositoryDispatch(ctx context.Context, client *github.Client, target Target, payload *WebhookPayload) error {

	owner := payload.Repository.Owner.Login

	logger.Info("sending repository dispatch",
		zap.String("target", fmt.Sprintf("%s/%s", owner, target.Repo)),
		zap.String("eventType", target.EventType),
	)

	// Build the client payload
	sourceEvent, _ := determineEventType(payload) 
	clientPayload := map[string]interface{}{
		"source_repo":  payload.Repository.FullName,
		"source_event": sourceEvent,
		"sender":       payload.Sender.Login,
	}

	// Add release info if available
	if payload.Release != nil {
		clientPayload["release"] = map[string]interface{}{
			"tag_name": payload.Release.TagName,
			"name":     payload.Release.Name,
			"draft":    payload.Release.Draft,
		}
	}
	
	payloadBytes, err := json.Marshal(clientPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	rawPayload := json.RawMessage(payloadBytes)

	_, _, err = client.Repositories.Dispatch(ctx, owner, target.Repo, github.DispatchRequestOptions{
		EventType:     target.EventType,
		ClientPayload: &rawPayload,
	})
	if err != nil {
		return fmt.Errorf("failed to dispatch: %w", err)
	}

	logger.Info("repository dispatch sent successfully",
		zap.String("target", fmt.Sprintf("%s/%s", owner, target.Repo)),
		zap.String("eventType", target.EventType),
	)

	return nil
}
