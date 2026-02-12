package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"go.uber.org/zap"
)

var logger *zap.Logger

func init() {
	var err error
	logger, err = zap.NewProduction()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	ctx := context.Background()
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Fatal("failed to load AWS config", zap.Error(err))
	}

	ssmClient = ssm.NewFromConfig(cfg)

	// Load GitHub App ID from SSM
	ssmAppIDPath := os.Getenv("SSM_GITHUB_APP_ID")
	if ssmAppIDPath != "" {
		if value, err := loadSSMParameter(ctx, ssmAppIDPath); err != nil {
			logger.Warn("failed to load GitHub App ID from SSM", zap.Error(err))
		} else {
			if appID, err := strconv.ParseInt(value, 10, 64); err == nil {
				githubAppID = appID
				logger.Info("GitHub App ID loaded from SSM", zap.Int64("appId", githubAppID))
			} else {
				logger.Warn("failed to parse GitHub App ID", zap.Error(err))
			}
		}
	}

	ssmGithubKeyPath := os.Getenv("SSM_GITHUB_APP_PRIVATE_KEY")
	ssmWebhookSecretPath := os.Getenv("SSM_GITHUB_APP_WEBHOOK_SECRET")

	if ssmGithubKeyPath != "" {
		if value, err := loadSSMParameter(ctx, ssmGithubKeyPath); err != nil {
			logger.Warn("failed to load GitHub App private key from SSM", zap.Error(err))
		} else {
			githubAppPrivateKeyPem = value
			logger.Info("GitHub App private key loaded from SSM")
		}
	}

	if ssmWebhookSecretPath != "" {
		if value, err := loadSSMParameter(ctx, ssmWebhookSecretPath); err != nil {
			logger.Warn("failed to load webhook secret from SSM", zap.Error(err))
		} else {
			githubAppWebhookSecret = value
			logger.Info("Webhook secret loaded from SSM")
			//Log the loaded webhook secret for debugging
			logger.Info("webhook secret value (REMOVE THIS LOG!)",
				zap.String("secret", githubAppWebhookSecret),
				zap.Int("length", len(githubAppWebhookSecret)),
			)
		}
	}
}

func handler(ctx context.Context, request events.LambdaFunctionURLRequest) (events.LambdaFunctionURLResponse, error) {
	logger.Info("received request",
		zap.String("requestId", request.RequestContext.RequestID),
		zap.String("method", request.RequestContext.HTTP.Method),
		zap.String("path", request.RawPath),
		zap.String("sourceIp", request.RequestContext.HTTP.SourceIP),
		zap.String("userAgent", request.RequestContext.HTTP.UserAgent),
	)

	// Debug: Log all headers to see what's actually in the request
	logger.Info("request headers received",
		zap.String("requestId", request.RequestContext.RequestID),
		zap.Any("headers", request.Headers),
	)

	// Debug: Log body size and preview (first 500 chars)
	bodyPreview := request.Body
	if len(bodyPreview) > 500 {
		bodyPreview = bodyPreview[:500] + "... (truncated)"
	}
	logger.Info("request body info",
		zap.String("requestId", request.RequestContext.RequestID),
		zap.Int("bodyLength", len(request.Body)),
		zap.String("bodyPreview", bodyPreview),
	)

	// AWS Lambda Function URLs normalize headers to lowercase
	signature := request.Headers["x-hub-signature-256"]
	if signature == "" {
		logger.Warn("missing x-hub-signature-256 header",
			zap.Any("availableHeaders", request.Headers),
		)
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       "Error: missing signature header",
		}, nil
	}

	// Verify GitHub webhook signature
	logger.Info("verifying signature", zap.String("signature", signature))
	valid, err := verifyGitHubSignature([]byte(request.Body), signature)
	if err != nil || !valid {
		logger.Warn("invalid signature", zap.Error(err))
		return events.LambdaFunctionURLResponse{
			StatusCode: 401,
			Body:       "Error: invalid signature",
		}, nil
	}

	// Parse the webhook payload
	var webhookPayload WebhookPayload
	if err := json.Unmarshal([]byte(request.Body), &webhookPayload); err != nil {
		logger.Error("failed to parse webhook payload", zap.Error(err))
		return events.LambdaFunctionURLResponse{
			StatusCode: 400,
			Body:       "Error: invalid payload",
		}, nil
	}

	// Validate event type before processing
	eventType, err := determineEventType(&webhookPayload)
	if err != nil {
		logger.Warn("unsupported event type, skipping",
			zap.Error(err),
			zap.String("repo", webhookPayload.Repository.FullName),
		)
		return events.LambdaFunctionURLResponse{
			StatusCode: 200,
			Body:       "Event type not supported",
		}, nil
	}

	logger.Info("webhook signature verified, processing event",
		zap.String("event", eventType),
		zap.String("repo", webhookPayload.Repository.FullName),
	)

	// Process the webhook and send repository dispatches
	if err := processWebhook(ctx, &webhookPayload); err != nil {
		logger.Error("failed to process webhook", zap.Error(err))
		return events.LambdaFunctionURLResponse{
			StatusCode: 500,
			Body:       "Error: failed to process webhook",
		}, nil
	}

	logger.Info("webhook processed successfully",
		zap.String("requestId", request.RequestContext.RequestID),
	)

	return events.LambdaFunctionURLResponse{
		StatusCode: 200,
		Body:       "Webhook received, verified, and processed",
	}, nil

}

func main() {
	defer logger.Sync()
	lambda.Start(handler)
}
