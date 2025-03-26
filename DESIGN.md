# Ollama-Claude Proxy Design Document

## Overview

This proxy allows Ollama clients to use Anthropic's Claude API. It translates requests from Ollama's format to Claude's format, forwards them to the Anthropic API, and then translates the responses back to Ollama's format.

## Architecture

```
+----------------+         +------------------+         +---------------+
|                |         |                  |         |               |
| Ollama Client  | ------> | Ollama-Claude    | ------> | Anthropic     |
| (curl, app)    |         | Proxy            |         | Claude API    |
|                | <------ | (Go Server)      | <------ |               |
+----------------+         +------------------+         +---------------+
```

## Components

1. **HTTP Server**: Listens for incoming requests on the `/generate` endpoint
2. **Request Translator**: Converts Ollama JSON format to Claude API format
3. **API Client**: Forwards requests to the Anthropic API with proper authentication
4. **Response Translator**: Converts Claude API responses back to Ollama format
5. **Configuration**: Loads API keys and settings from environment variables

## API Mapping

### Request Mapping

| Ollama Parameter        | Claude Parameter        | Notes                           |
|------------------------|--------------------------|--------------------------------|
| `model`                | `model`                  | Mapped to valid Claude model ID |
| `prompt`               | `prompt`                 | Wrapped with "Human: ... Assistant:" |
| `options.temperature`  | `temperature`            | Direct mapping                   |
| `options.top_p`        | `top_p`                  | Direct mapping                   |
| `options.top_k`        | `top_k`                  | Direct mapping                   |
| `options.num_predict`  | `max_tokens_to_sample`   | Max output tokens                |
| `stream`               | `stream`                 | Currently only support `false`   |

### Response Mapping

| Claude Parameter | Ollama Parameter | Notes                       |
|-----------------|-----------------|-----------------------------|
| `completion`     | `response`        | The generated text response  |
| N/A              | `model`           | Echo back the requested model |
| N/A              | `created_at`      | Current timestamp            |
| N/A              | `done`            | Always `true` for non-stream |

## Model Name Mapping

The proxy provides simple aliases for Claude models:

| Ollama Name       | Claude Model ID                 |
|------------------|--------------------------------|
| `claude`          | `claude-3-5-sonnet-20241022`    |
| `claude-3-sonnet` | `claude-3-sonnet-20240229`      |
| `claude-3-opus`   | `claude-3-opus-20240229`        |
| `claude-3-haiku`  | `claude-3-haiku-20240307`       |

## Key Considerations

1. **API Authentication**: Requires Anthropic API key in environment variable
2. **Prompt Formatting**: Prepends "Human:" and appends "Assistant:" to prompts
3. **Error Handling**: Proper forwarding of Claude API errors to clients
4. **Model Compatibility**: Simple mapping between model names
5. **Headers**: Includes required `anthropic-version` header

## Future Enhancements

1. Add support for streaming responses
2. Implement the Messages API for chat interfaces
3. Add more detailed metrics and logging
4. Support more Ollama parameters and endpoints
5. Add authentication for the proxy itself

## Security Considerations

1. API keys are loaded from environment variables, not hardcoded
2. No storage of user data or prompts
3. Consider rate limiting to prevent API abuse
4. Input validation to prevent malformed requests

## References

- [Anthropic Claude API Documentation](https://docs.anthropic.com/claude/reference/getting-started-with-the-api)
- [Ollama API Documentation](https://github.com/ollama/ollama/blob/main/docs/api.md)
