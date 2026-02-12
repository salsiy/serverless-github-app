package main

import (
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestConfigParsing(t *testing.T) {
	tests := []struct {
		name    string
		yaml    string
		wantErr bool
		verify  func(*testing.T, *AppConfig)
	}{
		{
			name: "valid config with single dispatch",
			yaml: `
dispatches:
  - event: release
    targets:
      - repo: target-repo
        event_type: deploy
`,
			wantErr: false,
			verify: func(t *testing.T, config *AppConfig) {
				if len(config.Dispatches) != 1 {
					t.Fatalf("expected 1 dispatch, got %d", len(config.Dispatches))
				}
				if config.Dispatches[0].Event != "release" {
					t.Errorf("event = %v, want release", config.Dispatches[0].Event)
				}
				if len(config.Dispatches[0].Targets) != 1 {
					t.Fatalf("expected 1 target, got %d", len(config.Dispatches[0].Targets))
				}
				if config.Dispatches[0].Targets[0].Repo != "target-repo" {
					t.Errorf("repo = %v, want target-repo", config.Dispatches[0].Targets[0].Repo)
				}
			},
		},
		{
			name: "multiple dispatches",
			yaml: `
dispatches:
  - event: release
    targets:
      - repo: repo1
        event_type: deploy
      - repo: repo2
        event_type: build
  - event: push
    targets:
      - repo: repo3
        event_type: test
`,
			wantErr: false,
			verify: func(t *testing.T, config *AppConfig) {
				if len(config.Dispatches) != 2 {
					t.Fatalf("expected 2 dispatches, got %d", len(config.Dispatches))
				}
				if len(config.Dispatches[0].Targets) != 2 {
					t.Errorf("expected 2 targets in first dispatch, got %d", len(config.Dispatches[0].Targets))
				}
				if len(config.Dispatches[1].Targets) != 1 {
					t.Errorf("expected 1 target in second dispatch, got %d", len(config.Dispatches[1].Targets))
				}
			},
		},
		{
			name: "empty config",
			yaml: `
dispatches: []
`,
			wantErr: false,
			verify: func(t *testing.T, config *AppConfig) {
				if len(config.Dispatches) != 0 {
					t.Errorf("expected 0 dispatches, got %d", len(config.Dispatches))
				}
			},
		},
		{
			name: "no dispatches key",
			yaml: `
other_key: value
`,
			wantErr: false,
			verify: func(t *testing.T, config *AppConfig) {
				if config.Dispatches != nil && len(config.Dispatches) != 0 {
					t.Errorf("expected nil or empty dispatches, got %d", len(config.Dispatches))
				}
			},
		},
		{
			name:    "invalid yaml",
			yaml:    `{invalid yaml: [}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := viper.New()
			v.SetConfigType("yaml")
			err := v.ReadConfig(strings.NewReader(tt.yaml))

			if tt.wantErr {
				if err == nil {
					t.Error("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error reading config: %v", err)
			}

			var config AppConfig
			err = v.Unmarshal(&config)
			if err != nil {
				t.Fatalf("unexpected error unmarshaling config: %v", err)
			}

			if tt.verify != nil {
				tt.verify(t, &config)
			}
		})
	}
}

func TestConfigEdgeCases(t *testing.T) {
	t.Run("config with empty target fields", func(t *testing.T) {
		yaml := `
dispatches:
  - event: release
    targets:
      - repo: ""
        event_type: ""
`
		v := viper.New()
		v.SetConfigType("yaml")
		err := v.ReadConfig(strings.NewReader(yaml))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var config AppConfig
		err = v.Unmarshal(&config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if config.Dispatches[0].Targets[0].Repo != "" {
			t.Error("expected empty repo")
		}
	})

	t.Run("config with special characters", func(t *testing.T) {
		yaml := `
dispatches:
  - event: "release-v2"
    targets:
      - repo: "org/repo-name"
        event_type: "deploy:production"
`
		v := viper.New()
		v.SetConfigType("yaml")
		err := v.ReadConfig(strings.NewReader(yaml))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var config AppConfig
		err = v.Unmarshal(&config)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if config.Dispatches[0].Event != "release-v2" {
			t.Errorf("event = %v, want release-v2", config.Dispatches[0].Event)
		}
	})
}
