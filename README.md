# Ollama-Claude Proxy

A proxy server that allows you to use Anthropic's Claude API with Ollama-compatible clients or directly using Claude's API format.

## Features

- **Direct Claude API Access**: Use the `/v1/messages` endpoint with Claude API format.
- **Ollama Compatibility**: Use the `/generate` endpoint with Ollama-style requests.
- **Model Mapping**: Simple names like `claude` are mapped to appropriate Claude model IDs.
- **Parameter Support**: Works with temperature, top_p, top_k, and other common settings.
- **Flexible Interface**: Supports both Ollama and Claude API clients.

## Setup

1. Clone this repository
2. Create an environment file with your Anthropic API key:
   ```bash
   cp .env.example .env
   # Edit .env to include your API key
   ```
3. Build the application:
   ```bash
   go build
   ```
4. Run the proxy with environment variables:
   ```bash
   source .env && ./ollama-claude-proxy
   ```

## Using with Claude API Format

You can send requests directly using Claude's API format:

```bash
curl -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [
      {
        "role": "user",
        "content": "What is the capital of France?"
      }
    ],
    "max_tokens": 100,
    "temperature": 0.7
  }'
```

## Using with Ollama Clients

With the proxy running, you can use Ollama clients by pointing them to your proxy URL (default: `http://localhost:8080`) instead of the Ollama server URL.

Example curl request:

```bash
curl -X POST http://localhost:8080/generate \
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

Environment variables:

- `ANTHROPIC_API_KEY`: Your Anthropic API key (required)
- `PORT`: Port to run the server on (default: 8080)

## Limitations

- Currently streams are not supported
- Some Ollama-specific features may not have Claude equivalents
- Claude has its own safety filters and policies that may differ from Ollama's

## License

MIT