.PHONY: build run test clean helm-package helm-install helm-uninstall

# Docker image configuration
IMAGE_NAME ?= ollama-claude-proxy
IMAGE_TAG ?= latest
DOCKER_REPO ?= 

# Helm configuration
HELM_RELEASE ?= ollama-claude-proxy
HELM_NAMESPACE ?= default

# Build the Docker image
build:
	docker build -t $(IMAGE_NAME):$(IMAGE_TAG) .

# Tag Docker image with repository
tag:
	docker tag $(IMAGE_NAME):$(IMAGE_TAG) $(DOCKER_REPO)/$(IMAGE_NAME):$(IMAGE_TAG)

# Push the Docker image to a registry
push: tag
	docker push $(DOCKER_REPO)/$(IMAGE_NAME):$(IMAGE_TAG)

# Run the application locally
run:
	go build -o ollama-claude-proxy
	./ollama-claude-proxy

# Run the Docker container locally
docker-run:
	docker run -p 8080:8080 --env-file .env $(IMAGE_NAME):$(IMAGE_TAG)

# Run tests
test:
	go test -v ./...

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

# Clean build artifacts
clean:
	rm -f ollama-claude-proxy
	rm -f coverage.out

# Package the Helm chart
helm-package:
	helm package helm/ollama-claude-proxy

# Install the Helm chart
helm-install:
	helm install $(HELM_RELEASE) helm/ollama-claude-proxy \
		--namespace $(HELM_NAMESPACE) \
		--set image.repository=$(DOCKER_REPO)/$(IMAGE_NAME) \
		--set image.tag=$(IMAGE_TAG)

# Uninstall the Helm chart
helm-uninstall:
	helm uninstall $(HELM_RELEASE) --namespace $(HELM_NAMESPACE)