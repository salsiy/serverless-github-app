package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// verifyGitHubSignature verifies that the webhook request is from GitHub
// by comparing the signature in the X-Hub-Signature-256 header with
// https://docs.github.com/en/webhooks/webhook-events-and-payloads
// the HMAC-SHA256 hash of the request body using the webhook secret

func verifyGitHubSignature(payload []byte, signature string) (bool, error) {
	if githubAppWebhookSecret == "" {
		return false, fmt.Errorf("webhook secret not configured")
	}

	if signature == "" {
		return false, fmt.Errorf("no signature provided")
	}

	if !strings.HasPrefix(signature, "sha256=") {
		return false, fmt.Errorf("invalid signature format")
	}

	receivedSignature := strings.TrimPrefix(signature, "sha256=")

	
	mac := hmac.New(sha256.New, []byte(githubAppWebhookSecret))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	if !hmac.Equal([]byte(expectedSignature), []byte(receivedSignature)) {
		return false, fmt.Errorf("signature mismatch")
	}
	return true, nil
}
