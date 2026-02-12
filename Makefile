.PHONY: deploy plan clean init test

init:
	@echo "Initializing Terraform..."
	cd infra && terraform init -upgrade

plan: init
	@echo "Planning Terraform changes..."
	cd infra && terraform plan

deploy: init
	@echo "Deploying to AWS..."
	@echo "Note: The terraform-aws-modules/lambda will automatically build the Go binary"
	cd infra && terraform apply -auto-approve
	@echo ""
	@echo "Deployment complete!"
	cd infra && terraform output

destroy:
	@echo "Destroying AWS resources..."
	cd infra && terraform destroy -auto-approve

clean:
	@echo "Cleaning artifacts..."
	rm -f app/bootstrap
	rm -rf infra/.terraform
	rm -f infra/.terraform.lock.hcl
	rm -rf infra/builds
	@echo "Clean complete!"

# Testing commands
test:
	@echo "Running tests..."
	cd app && go test -v ./...
