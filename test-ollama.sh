#!/bin/bash
# Test script for Ollama-Claude proxy - Ollama compatibility endpoint

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

# Test the Ollama compatibility endpoint (/generate)
echo "Testing Ollama compatibility endpoint (/generate)..."

RESPONSE=$(curl -s -X POST http://localhost:8080/generate \
  -H "Content-Type: application/json" \
  -d "{
    \"model\": \"claude-3-opus\",
    \"prompt\": \"$PROMPT\",
    \"options\": {
      \"temperature\": 0.7,
      \"top_p\": 0.95,
      \"top_k\": 40,
      \"num_predict\": 100
    },
    \"stream\": false
  }")

# Check if the response is valid JSON
if ! echo "$RESPONSE" | jq -e . >/dev/null 2>&1; then
  echo "Error: Invalid response received"
  echo "$RESPONSE"
  exit 1
fi

# Display the response text
echo "Claude's response via Ollama compatibility:"
echo "----------------"
echo "$RESPONSE" | jq -r '.response'
echo "----------------"

# Show full response in JSON format
echo
echo "Full JSON response:"
echo "$RESPONSE" | jq