# Serverless GitHub App

A serverless AWS Lambda function that triggers cross-repository workflows via GitHub App webhooks. It receives release events and dispatches `repository_dispatch` events to configured target repositories, enabling automated cross-repository CI/CD pipelines.

## Overview

This GitHub App:
1. Receives webhook events when releases are published
2. Loads configuration from `.github/app-config.yaml` in the source repository
3. Authenticates as a GitHub App installation
4. Sends `repository_dispatch` events to target repositories
5. Triggers GitHub Actions workflows in downstream repositories

## Key Dependencies

### GitHub Integration
- **[google/go-github](https://github.com/google/go-github)** (`v57.0.0`) - GitHub API client for reading configs and sending repository dispatches
- **[ghinstallation](https://github.com/bradleyfalzon/ghinstallation)** (`v2.17.0`) - GitHub App authentication (handles JWT and installation tokens)

### AWS Services
- **[aws-lambda-go](https://github.com/aws/aws-lambda-go)** (`v1.51.1`) - AWS Lambda runtime and handler
- **[aws-sdk-go-v2/ssm](https://github.com/aws/aws-sdk-go-v2)** (`v1.67.7`) - loading secrets from ssm

### Configuration & Logging
- **[viper](https://github.com/spf13/viper)** (`v1.21.0`) - YAML config parsing

### How They Work Together

1. **Lambda receives webhook** → `aws-lambda-go` handles the request
2. **Load secrets** → `aws-sdk-go-v2/ssm` retrieves credentials from Parameter Store
3. **Authenticate with GitHub** → `ghinstallation` creates authenticated client using App credentials
4. **Fetch config** → `go-github` reads `.github/app-config.yaml` from repository
5. **Parse config** → `viper` unmarshals YAML into Go structs
6. **Send dispatches** → `go-github` calls repository_dispatch API
7. **Log everything** → `zap` outputs structured JSON logs to CloudWatch

## Prerequisites

1. **GitHub App** with:
   - Permissions: Repository contents (Read), Repository administration (Read & Write)
   - Webhook events: `release`
   - Webhook secret generated
   
2. **AWS Account** with credentials configured
3. **Terraform** (`>= 1.5`)
4. **Go** (`1.24.4`)

## Setup

### 1. Store GitHub App Credentials in AWS SSM

```bash
# GitHub App ID (from your GitHub App settings)
aws ssm put-parameter \
  --name "/dev/github-app/app-id" \
  --value "YOUR_APP_ID" \
  --type "String"

# GitHub App Private Key (download from GitHub App settings)
aws ssm put-parameter \
  --name "/dev/github-app-private-key" \
  --value file://your-app.private-key.pem \
  --type "SecureString"

# Webhook Secret (set in GitHub App settings)
aws ssm put-parameter \
  --name "/dev/github-app-webhook-secret" \
  --value "YOUR_WEBHOOK_SECRET" \
  --type "SecureString"
```

### 2. Deploy Infrastructure

```bash
# Deploy to AWS
make deploy

# Copy the Lambda Function URL from the output
```

### 3. Configure GitHub App Webhook

1. Go to your GitHub App settings
2. Set **Webhook URL** to your Lambda Function URL
3. Set **Webhook secret** (same value used in SSM)
4. Enable **Release** events
5. Save configuration

## Configuration

### Source Repository (`.github/app-config.yaml`)

Create this file in repositories that will **trigger** dispatches:

```yaml
dispatches:
  - event: "release"
    targets:
      - repo: "target-repo-1"
        event_type: "deploy"
      
      - repo: "another-repo"
        event_type: "upstream-release"
```

**Fields:**
- `event` - Event type (`"release"` only currently supported)
- `repo` - Target repository name (must be in same organization)
- `event_type` - Custom event type for repository_dispatch

### Target Repository Workflow

Create a workflow to receive dispatches:

```yaml
name: Deploy on Repository Dispatch

on:
  repository_dispatch:
    types: [deploy]  # Match event_type from config

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Echo event data
        run: |
          echo "Source: ${{ github.event.client_payload.source_repo }}"
          echo "Tag: ${{ github.event.client_payload.release.tag_name }}"
      
      # Add deployment steps here
```

**Client Payload:**
```json
{
  "source_repo": "owner/repo-name",
  "source_event": "release",
  "sender": "username",
  "release": {
    "tag_name": "v1.0.0",
    "name": "Release Name",
    "draft": false
  }
}
```


## Monitoring

View logs in CloudWatch:
```bash
aws logs tail /aws/lambda/serverless-github-app-123456789012-us-east-1 --follow
```

## Testing

Run unit tests:
```bash
make test
```


