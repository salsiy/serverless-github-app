package main

import (
	"testing"
)

func TestDetermineEventType(t *testing.T) {
	tests := []struct {
		name        string
		payload     *WebhookPayload
		wantEvent   string
		wantErr     bool
		description string
	}{
		{
			name: "release event",
			payload: &WebhookPayload{
				Action: "published",
				Release: &Release{
					ID:      123,
					TagName: "v1.0.0",
					Name:    "Release 1.0.0",
					Draft:   false,
				},
			},
			wantEvent:   "release",
			wantErr:     false,
			description: "should detect release event from payload",
		},
		{
			name: "no release",
			payload: &WebhookPayload{
				Action:  "created",
				Release: nil,
			},
			wantEvent:   "",
			wantErr:     true,
			description: "should return error when no release present",
		},
		{
			name: "release with draft",
			payload: &WebhookPayload{
				Action: "created",
				Release: &Release{
					ID:      456,
					TagName: "v2.0.0",
					Name:    "Release 2.0.0",
					Draft:   true,
				},
			},
			wantEvent:   "release",
			wantErr:     false,
			description: "should detect release event even if draft",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eventType, err := determineEventType(tt.payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("determineEventType() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if eventType != tt.wantEvent {
				t.Errorf("determineEventType() = %v, want %v", eventType, tt.wantEvent)
			}
		})
	}
}

func TestMatchesRule(t *testing.T) {
	tests := []struct {
		name      string
		rule      Rule
		eventType string
		want      bool
	}{
		{
			name: "matching release rule",
			rule: Rule{
				Event: "release",
				Targets: []Target{
					{Repo: "target-repo", EventType: "deploy"},
				},
			},
			eventType: "release",
			want:      true,
		},
		{
			name: "non-matching rule",
			rule: Rule{
				Event: "push",
				Targets: []Target{
					{Repo: "target-repo", EventType: "deploy"},
				},
			},
			eventType: "release",
			want:      false,
		},
		{
			name: "empty event type",
			rule: Rule{
				Event: "",
				Targets: []Target{
					{Repo: "target-repo", EventType: "deploy"},
				},
			},
			eventType: "release",
			want:      false,
		},
		{
			name: "empty rule event",
			rule: Rule{
				Event: "release",
				Targets: []Target{
					{Repo: "target-repo", EventType: "deploy"},
				},
			},
			eventType: "",
			want:      false,
		},
		{
			name: "case sensitive match",
			rule: Rule{
				Event: "Release",
				Targets: []Target{
					{Repo: "target-repo", EventType: "deploy"},
				},
			},
			eventType: "release",
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesRule(tt.rule, tt.eventType)
			if got != tt.want {
				t.Errorf("matchesRule() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSupportedEvents(t *testing.T) {
	t.Run("release is supported", func(t *testing.T) {
		if !supportedEvents["release"] {
			t.Error("release event should be supported")
		}
	})

	t.Run("unsupported events", func(t *testing.T) {
		unsupported := []string{"push", "pull_request", "issues", "fork", "star"}
		for _, event := range unsupported {
			if supportedEvents[event] {
				t.Errorf("event %s should not be supported", event)
			}
		}
	})
}
