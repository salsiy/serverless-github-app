package main

import (
	"encoding/json"
	"testing"
)

func TestWebhookPayloadUnmarshal(t *testing.T) {
	tests := []struct {
		name    string
		json    string
		wantErr bool
		verify  func(*testing.T, *WebhookPayload)
	}{
		{
			name: "complete release payload",
			json: `{
				"action": "published",
				"repository": {
					"id": 123456,
					"name": "test-repo",
					"full_name": "owner/test-repo",
					"owner": {
						"login": "owner",
						"id": 789
					}
				},
				"sender": {
					"login": "user",
					"id": 456
				},
				"installation": {
					"id": 999
				},
				"release": {
					"id": 111,
					"tag_name": "v1.0.0",
					"name": "Release 1.0.0",
					"draft": false
				}
			}`,
			wantErr: false,
			verify: func(t *testing.T, p *WebhookPayload) {
				if p.Action != "published" {
					t.Errorf("Action = %v, want published", p.Action)
				}
				if p.Repository.Name != "test-repo" {
					t.Errorf("Repository.Name = %v, want test-repo", p.Repository.Name)
				}
				if p.Release == nil {
					t.Fatal("Release should not be nil")
				}
				if p.Release.TagName != "v1.0.0" {
					t.Errorf("Release.TagName = %v, want v1.0.0", p.Release.TagName)
				}
			},
		},
		{
			name: "payload without release",
			json: `{
				"action": "created",
				"repository": {
					"id": 123456,
					"name": "test-repo",
					"full_name": "owner/test-repo",
					"owner": {
						"login": "owner",
						"id": 789
					}
				},
				"sender": {
					"login": "user",
					"id": 456
				},
				"installation": {
					"id": 999
				}
			}`,
			wantErr: false,
			verify: func(t *testing.T, p *WebhookPayload) {
				if p.Release != nil {
					t.Error("Release should be nil")
				}
				if p.Action != "created" {
					t.Errorf("Action = %v, want created", p.Action)
				}
			},
		},
		{
			name: "minimal valid payload",
			json: `{
				"repository": {
					"owner": {}
				},
				"sender": {},
				"installation": {}
			}`,
			wantErr: false,
			verify: func(t *testing.T, p *WebhookPayload) {
				if p.Repository.Owner.Login != "" {
					t.Errorf("Owner.Login should be empty, got %v", p.Repository.Owner.Login)
				}
			},
		},
		{
			name:    "invalid json",
			json:    `{invalid json}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var payload WebhookPayload
			err := json.Unmarshal([]byte(tt.json), &payload)

			if (err != nil) != tt.wantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.verify != nil {
				tt.verify(t, &payload)
			}
		})
	}
}

func TestAppConfigStructure(t *testing.T) {
	config := AppConfig{
		Dispatches: []Rule{
			{
				Event: "release",
				Targets: []Target{
					{Repo: "repo1", EventType: "deploy"},
					{Repo: "repo2", EventType: "build"},
				},
			},
		},
	}

	if len(config.Dispatches) != 1 {
		t.Errorf("Expected 1 dispatch rule, got %d", len(config.Dispatches))
	}

	rule := config.Dispatches[0]
	if rule.Event != "release" {
		t.Errorf("Expected event 'release', got '%s'", rule.Event)
	}

	if len(rule.Targets) != 2 {
		t.Errorf("Expected 2 targets, got %d", len(rule.Targets))
	}

	if rule.Targets[0].Repo != "repo1" {
		t.Errorf("Expected repo 'repo1', got '%s'", rule.Targets[0].Repo)
	}
}

func TestReleaseStructure(t *testing.T) {
	release := Release{
		ID:      12345,
		TagName: "v1.2.3",
		Name:    "Version 1.2.3",
		Draft:   true,
	}

	if release.TagName != "v1.2.3" {
		t.Errorf("TagName = %v, want v1.2.3", release.TagName)
	}

	if !release.Draft {
		t.Error("Draft should be true")
	}
}
