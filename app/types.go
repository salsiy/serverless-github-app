package main

type AppConfig struct {
	Dispatches []Rule `yaml:"dispatches" mapstructure:"dispatches"`
}

type Rule struct {
	Event   string   `yaml:"event" mapstructure:"event"`
	Targets []Target `yaml:"targets" mapstructure:"targets"`
}

type Target struct {
	Repo      string `yaml:"repo" mapstructure:"repo"`
	EventType string `yaml:"event_type" mapstructure:"event_type"`
}

// WebhookPayload represents the GitHub webhook payload
type WebhookPayload struct {
	Action       string       `json:"action"`
	Repository   Repository   `json:"repository"`
	Sender       User         `json:"sender"`
	Installation Installation `json:"installation"`
	Release      *Release     `json:"release,omitempty"`
	Ref          string       `json:"ref,omitempty"`
}

type Repository struct {
	ID       int64  `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	Owner    User   `json:"owner"`
}

type User struct {
	Login string `json:"login"`
	ID    int64  `json:"id"`
}

type Installation struct {
	ID int64 `json:"id"`
}

type Release struct {
	ID      int64  `json:"id"`
	TagName string `json:"tag_name"`
	Name    string `json:"name"`
	Draft   bool   `json:"draft"`
}
