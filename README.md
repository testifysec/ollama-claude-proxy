# Ollama-Claude Proxy

A proxy server that allows you to use Anthropic's Claude API with Ollama-compatible clients or directly using Claude's API format.

## Features

- **Ollama Compatibility**: Use the `/api/generate` endpoint with Ollama-style requests.
- **Built-in Testing UI**: Use the web interface at the root URL to test the proxy.
- **Model Mapping**: Simple names like `claude` are mapped to appropriate Claude model IDs.
- **Parameter Support**: Works with temperature, top_p, top_k, and other common settings.
- **Flexible Interface**: Supports both Ollama and Claude API clients.
- **Docker Support**: Run as a container with the provided Dockerfile.
- **Kubernetes Support**: Deploy to Kubernetes using the included Helm chart.

## Quick Setup

1. Clone this repository
2. Create an environment file with your Anthropic API key:
   ```bash
   cp .env.example .env
   # Edit .env to include your API key
   ```
3. Build and run:
   ```bash
   # Local build
   go build
   source .env && ./ollama-claude-proxy
   
   # Or using make
   make run
   ```

## Using the Testing UI

Once the server is running, you can access the built-in testing UI by opening your browser to:

```
http://localhost:8080/
```

The UI provides a simple interface to:
- Select a Claude model to use
- Adjust parameters like temperature and max tokens
- Enter a prompt
- View the generated response
- See the raw JSON request being sent

![Claude API Tester UI Screenshot](docs/images/ui-screenshot.png)
(You'll need to create this screenshot after running the server)

## Using with Ollama Clients

With the proxy running, you can use Ollama clients by pointing them to your proxy URL (default: `http://localhost:8080`) instead of the Ollama server URL.

Example curl request:

```bash
curl -X POST http://localhost:8080/api/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "prompt": "What is the capital of France?",
    "options": {
      "temperature": 0.7,
      "top_p": 0.95,
      "top_k": 40,
      "num_predict": 100
    },
    "stream": false
  }'
```

## Model Mapping

The proxy maps simple model names to Claude model IDs:

- `claude` → `claude-3-opus-20240229`
- `claude-3-opus` → `claude-3-opus-20240229`
- `claude-3-sonnet` → `claude-3-sonnet-20240229`
- `claude-3-haiku` → `claude-3-haiku-20240307`
- `claude-3.5-sonnet` → `claude-3-5-sonnet-latest`
- `claude-3.7-sonnet` → `claude-3-7-sonnet-latest`
- `claude-2.1` → `claude-2.1`

## Configuration

### Environment Variables

- `ANTHROPIC_API_KEY`: Your Anthropic API key (required)
- `PORT`: Port to run the server on (default: 8080)

### Config File

You can also provide configuration via a JSON file using the `-config` flag:

```bash
./ollama-claude-proxy -config=/path/to/config.json
```

Example config.json:
```json
{
  "port": "8080",
  "api_key": "sk-ant-your-api-key-here",
  "api_version": "2023-06-01",
  "api_endpoint": "https://api.anthropic.com/v1/messages",
  "system_prompt": "You are Claude, an AI assistant by Anthropic.",
  "default_model": "claude-3-5-sonnet-20240620",
  "request_timeout_secs": 60
}
```

Note: The API key can be provided either in the config file or as an environment variable. For security reasons, using the environment variable is recommended for local development.

## Docker Support

Build and run the Docker image:

```bash
# Build the image
docker build -t ollama-claude-proxy .

# Run the container
docker run -p 8080:8080 -e ANTHROPIC_API_KEY=sk-ant-your-api-key-here ollama-claude-proxy
```

Or using Make:

```bash
# Build image
make build

# Run container (uses .env file)
make docker-run
```

## Kubernetes Deployment

The project includes a Helm chart for easy deployment to Kubernetes.

```bash
# Install using Helm
helm install ollama-claude-proxy ./helm/ollama-claude-proxy \
  --set secret.apiKey=sk-ant-your-api-key-here

# Or to override configuration:
helm install ollama-claude-proxy ./helm/ollama-claude-proxy \
  --set secret.apiKey=sk-ant-your-api-key-here \
  --set config.systemPrompt="You are Claude, a helpful AI assistant." \
  --set config.defaultModel=claude-3-opus-20240229

# Or using the Makefile (set DOCKER_REPO first)
DOCKER_REPO=your-docker-repo make helm-install
```

The Helm chart uses a ConfigMap for configuration and a Secret for storing the API key.

### Minikube Deployment

For local testing and development, you can use Minikube to deploy the proxy:

```bash
# Start Minikube if not already running
minikube start

# Deploy to Minikube using the script
source .env  # To load your ANTHROPIC_API_KEY
make minikube-deploy

# Run tests against the Minikube deployment
make minikube-test
make test-client-minikube

# Delete the deployment when done
make minikube-delete
```

The deployment script will:
1. Build a Docker image in Minikube's Docker environment
2. Create a Kubernetes namespace if needed
3. Deploy the Helm chart to the Minikube cluster
4. Run validation tests to ensure the proxy is working correctly
5. Output the URL to access the service

For more details on Helm chart configuration, see [helm/ollama-claude-proxy/README.md](helm/ollama-claude-proxy/README.md).

## Limitations

- Currently streams are not supported
- Some Ollama-specific features may not have Claude equivalents
- Claude has its own safety filters and policies that may differ from Ollama's

## License

MIT