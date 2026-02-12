package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"
)

func TestVerifyGitHubSignature(t *testing.T) {
	testSecret := "test-webhook-secret"
	originalSecret := githubAppWebhookSecret
	githubAppWebhookSecret = testSecret
	defer func() { githubAppWebhookSecret = originalSecret }()

	payload := []byte(`{"action":"published","repository":{"full_name":"test/repo"}}`)

	// Generate valid signature
	mac := hmac.New(sha256.New, []byte(testSecret))
	mac.Write(payload)
	validSignature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		name          string
		payload       []byte
		signature     string
		secret        string
		wantValid     bool
		wantErr       bool
		expectedError string
	}{
		{
			name:      "valid signature",
			payload:   payload,
			signature: validSignature,
			secret:    testSecret,
			wantValid: true,
			wantErr:   false,
		},
		{
			name:          "invalid signature",
			payload:       payload,
			signature:     "sha256=invalid",
			secret:        testSecret,
			wantValid:     false,
			wantErr:       true,
			expectedError: "signature mismatch",
		},
		{
			name:          "missing sha256 prefix",
			payload:       payload,
			signature:     "invalid",
			secret:        testSecret,
			wantValid:     false,
			wantErr:       true,
			expectedError: "invalid signature format",
		},
		{
			name:          "empty signature",
			payload:       payload,
			signature:     "",
			secret:        testSecret,
			wantValid:     false,
			wantErr:       true,
			expectedError: "no signature provided",
		},
		{
			name:          "empty secret",
			payload:       payload,
			signature:     validSignature,
			secret:        "",
			wantValid:     false,
			wantErr:       true,
			expectedError: "webhook secret not configured",
		},
		{
			name:      "different payload should fail",
			payload:   []byte(`{"different":"payload"}`),
			signature: validSignature,
			secret:    testSecret,
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			githubAppWebhookSecret = tt.secret

			valid, err := verifyGitHubSignature(tt.payload, tt.signature)

			if (err != nil) != tt.wantErr {
				t.Errorf("verifyGitHubSignature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if valid != tt.wantValid {
				t.Errorf("verifyGitHubSignature() valid = %v, want %v", valid, tt.wantValid)
			}

			if tt.wantErr && tt.expectedError != "" && err != nil {
				if err.Error() != tt.expectedError {
					t.Errorf("verifyGitHubSignature() error = %v, want %v", err.Error(), tt.expectedError)
				}
			}
		})
	}
}

func TestVerifyGitHubSignature_EdgeCases(t *testing.T) {
	originalSecret := githubAppWebhookSecret
	defer func() { githubAppWebhookSecret = originalSecret }()

	githubAppWebhookSecret = "test-secret"

	t.Run("empty payload", func(t *testing.T) {
		payload := []byte{}
		mac := hmac.New(sha256.New, []byte(githubAppWebhookSecret))
		mac.Write(payload)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		valid, err := verifyGitHubSignature(payload, signature)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !valid {
			t.Error("expected valid signature for empty payload")
		}
	})

	t.Run("large payload", func(t *testing.T) {
		payload := make([]byte, 1024*1024) // 1MB
		for i := range payload {
			payload[i] = byte(i % 256)
		}

		mac := hmac.New(sha256.New, []byte(githubAppWebhookSecret))
		mac.Write(payload)
		signature := "sha256=" + hex.EncodeToString(mac.Sum(nil))

		valid, err := verifyGitHubSignature(payload, signature)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if !valid {
			t.Error("expected valid signature for large payload")
		}
	})
}
