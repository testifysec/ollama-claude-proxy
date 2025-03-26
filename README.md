# Ollama-Claude Proxy

A simple proxy server that allows you to use Anthropic's Claude API with Ollama-compatible clients. This proxy translates Ollama API requests to Claude API requests, allowing you to leverage Claude's powerful AI capabilities with tools built for Ollama.

## Features

- Accepts Ollama-compatible `/generate` endpoint requests
- Forwards requests to Claude's API using your API key
- Returns responses in Ollama-compatible format
- Supports common generation parameters (temperature, top_p, top_k, etc.)
- Simple model name mapping ("claude" → appropriate Claude model)

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

## Usage

With the proxy running, you can use Ollama clients by pointing them to your proxy URL (default: `http://localhost:8080`) instead of the Ollama server URL.

Example curl request:

```bash
curl -X POST http://localhost:8080/generate -d '{
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

- `claude` → `claude-3-5-sonnet-20241022`
- `claude-3-sonnet` → `claude-3-sonnet-20240229`
- `claude-3-opus` → `claude-3-opus-20240229`
- `claude-3-haiku` → `claude-3-haiku-20240307`

You can also use the full model ID directly.

## Configuration

Environment variables:

- `ANTHROPIC_API_KEY`: Your Anthropic API key (required)
- `PORT`: Port to run the server on (default: 8080)

## Limitations

- Currently supports non-streaming responses only (`stream: false`)
- Some Ollama-specific features may not have Claude equivalents
- Claude has its own safety filters and policies that may differ from Ollama's

## License

MIT
