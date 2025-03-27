#!/bin/bash
set -e

# This script tests the Ollama-Claude proxy API endpoints
# Usage: ./test-client.sh [base_url]
# If base_url is not provided, it defaults to http://localhost:8080

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Get base URL from command line or use default
BASE_URL=${1:-"http://localhost:8080"}

echo -e "${YELLOW}Testing Ollama-Claude proxy at $BASE_URL...${NC}"

# Test health endpoint
echo -e "\n${YELLOW}Testing health endpoint...${NC}"
HEALTH_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" $BASE_URL/health)
if [ "$HEALTH_RESPONSE" = "200" ]; then
    echo -e "${GREEN}Health check successful (200 OK)${NC}"
else
    echo -e "${RED}Health check failed with status code: $HEALTH_RESPONSE${NC}"
    exit 1
fi

# Test Claude API endpoint
echo -e "\n${YELLOW}Testing Claude Messages API endpoint...${NC}"
CLAUDE_RESPONSE=$(curl -s -X POST $BASE_URL/v1/messages \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-opus-20240229",
    "messages": [
      {
        "role": "user",
        "content": "Say hello"
      }
    ],
    "max_tokens": 10
  }')

echo "Response:"
echo "$CLAUDE_RESPONSE" | jq . 2>/dev/null || echo "$CLAUDE_RESPONSE"

# Test Ollama compatibility endpoint
echo -e "\n${YELLOW}Testing Ollama compatibility endpoint...${NC}"
OLLAMA_RESPONSE=$(curl -s -X POST $BASE_URL/generate \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude",
    "prompt": "Say hello",
    "options": {
      "temperature": 0.7,
      "num_predict": 10
    },
    "stream": false
  }')

echo "Response:"
echo "$OLLAMA_RESPONSE" | jq . 2>/dev/null || echo "$OLLAMA_RESPONSE"

echo -e "\n${GREEN}All tests completed!${NC}"