package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"go.uber.org/zap"
)

var (
	ssmClient              *ssm.Client
	githubAppPrivateKeyPem string
	githubAppWebhookSecret string
	githubAppID            int64 
)

func loadSSMParameter(ctx context.Context, paramName string) (string, error) {
	input := &ssm.GetParameterInput{
		Name:           &paramName,
		WithDecryption: aws.Bool(true),
	}

	result, err := ssmClient.GetParameter(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to get parameter: %w", err)
	}

	logger.Info("SSM parameter retrieved", zap.String("parameterName", paramName))

	return *result.Parameter.Value, nil
}
