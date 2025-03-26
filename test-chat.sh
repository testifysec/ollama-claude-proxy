#!/bin/bash
# Test script for Ollama-Claude proxy

# Check if API is running
echo "Testing connection to Ollama-Claude proxy..."
HEALTH_CHECK=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/health)

if [ "$HEALTH_CHECK" != "200" ]; then
  echo "Error: Proxy server not running or health check failed."
  echo "Please run the server with: source .env && ./ollama-claude-proxy"
  exit 1
fi

echo "Connection successful!"
echo

# Get prompt from arguments or use default
if [ $# -eq 0 ]; then
  PROMPT="What is the capital of France? Please keep your answer short."
else
  PROMPT="$*"
fi

echo "Sending request with prompt: '$PROMPT'"
echo

# Testing direct Claude API client format using /v1/messages
echo "Testing direct Claude API messages format..."

RESPONSE=$(curl -s -X POST http://localhost:8080/v1/messages \
  -H "Content-Type: application/json" \
  -H "x-api-key: demo-api-key" \
  -d "{
    \"model\": \"claude-3-opus-20240229\",
    \"messages\": [
      {
        \"role\": \"user\",
        \"content\": \"$PROMPT\"
      }
    ],
    \"max_tokens\": 100,
    \"temperature\": 0.7
  }")

# Check if the response is valid JSON
if ! echo "$RESPONSE" | jq -e . >/dev/null 2>&1; then
  echo "Error: Invalid response received"
  echo "$RESPONSE"
  exit 1
fi

# Display the response text
echo "Claude's response:"
echo "----------------"
echo "$RESPONSE" | jq -r '.content[0].text'
echo "----------------"

# Show full response in JSON format
echo
echo "Full JSON response:"
echo "$RESPONSE" | jq