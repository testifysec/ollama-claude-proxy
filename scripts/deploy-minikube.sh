#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Check if minikube is installed
if ! command -v minikube &> /dev/null; then
    echo -e "${RED}Minikube is not installed. Please install it first.${NC}"
    echo "Visit https://minikube.sigs.k8s.io/docs/start/"
    exit 1
fi

# Check if helm is installed
if ! command -v helm &> /dev/null; then
    echo -e "${RED}Helm is not installed. Please install it first.${NC}"
    echo "Visit https://helm.sh/docs/intro/install/"
    exit 1
fi

# Check if minikube is running
MINIKUBE_STATUS=$(minikube status --format '{{.Host}}' 2>/dev/null || echo "Not Running")
if [ "$MINIKUBE_STATUS" != "Running" ]; then
    echo -e "${YELLOW}Minikube is not running. Starting minikube...${NC}"
    minikube start
fi

# Set docker env to use minikube's docker daemon
echo -e "${YELLOW}Setting docker environment to use minikube...${NC}"
eval $(minikube docker-env)

# Build docker image
echo -e "${YELLOW}Building docker image...${NC}"
docker build -t ollama-claude-proxy:local .

# Check if namespace exists
NAMESPACE="ollama-claude-proxy"
if ! kubectl get namespace "$NAMESPACE" &> /dev/null; then
    echo -e "${YELLOW}Creating namespace $NAMESPACE...${NC}"
    kubectl create namespace "$NAMESPACE"
fi

# Check if the API key is in the environment
if [ -z "$ANTHROPIC_API_KEY" ]; then
    if [ -f "../.env" ]; then
        # Try to source the .env file
        echo -e "${YELLOW}ANTHROPIC_API_KEY not set, trying to source from ../.env file...${NC}"
        source "../.env"
    fi
    
    if [ -z "$ANTHROPIC_API_KEY" ]; then
        echo -e "${RED}ANTHROPIC_API_KEY environment variable is not set!${NC}"
        echo "Please set it before running this script:"
        echo "export ANTHROPIC_API_KEY=your-api-key"
        exit 1
    fi
fi

# Install or upgrade the helm chart
echo -e "${YELLOW}Deploying ollama-claude-proxy to minikube...${NC}"
# Use the absolute path to the helm chart
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
HELM_CHART_PATH="$SCRIPT_DIR/../helm/ollama-claude-proxy"
helm upgrade --install ollama-claude-proxy "$HELM_CHART_PATH" \
    --namespace "$NAMESPACE" \
    --set image.repository=ollama-claude-proxy \
    --set image.tag=local \
    --set image.pullPolicy=Never \
    --set env.ANTHROPIC_API_KEY="$ANTHROPIC_API_KEY" \
    --set service.type=NodePort

# Wait for the deployment to be ready
echo -e "${YELLOW}Waiting for deployment to be ready...${NC}"
kubectl rollout status deployment/ollama-claude-proxy -n "$NAMESPACE" --timeout=120s

# Get the URL to access the service
echo -e "${YELLOW}Getting service URL...${NC}"
MINIKUBE_IP=$(minikube ip)
NODE_PORT=$(kubectl get svc -n "$NAMESPACE" ollama-claude-proxy -o jsonpath='{.spec.ports[0].nodePort}')
SERVICE_URL="http://$MINIKUBE_IP:$NODE_PORT"

echo -e "${GREEN}Ollama-Claude proxy deployed successfully!${NC}"
echo -e "${GREEN}You can access the service at:${NC} $SERVICE_URL"
echo ""

# Run an internal validation test
echo -e "${YELLOW}Running internal validation tests...${NC}"
kubectl run --rm -i test-client --image=curlimages/curl:7.86.0 -n "$NAMESPACE" --restart=Never -- sh -c "
  echo 'Testing health endpoint...'
  HEALTH_CODE=\$(curl -s -o /dev/null -w '%{http_code}' http://ollama-claude-proxy:8080/health)
  if [ \"\$HEALTH_CODE\" != \"200\" ]; then
    echo \"Health check failed with status code \$HEALTH_CODE\"
    exit 1
  fi
  echo 'Health check successful'
  
  echo 'Testing Claude API endpoint...'
  RESPONSE=\$(curl -s -X POST http://ollama-claude-proxy:8080/v1/messages \\
    -H \"Content-Type: application/json\" \\
    -d '{
      \"model\": \"claude-3-opus-20240229\",
      \"messages\": [
        {
          \"role\": \"user\",
          \"content\": \"Say test\"
        }
      ],
      \"max_tokens\": 10
    }')
    
  echo \"\$RESPONSE\"
  
  echo 'Testing Ollama compatibility endpoint...'
  RESPONSE=\$(curl -s -X POST http://ollama-claude-proxy:8080/generate \\
    -H \"Content-Type: application/json\" \\
    -d '{
      \"model\": \"claude\",
      \"prompt\": \"Say test\",
      \"options\": {
        \"temperature\": 0.7,
        \"num_predict\": 10
      },
      \"stream\": false
    }')
    
  echo \"\$RESPONSE\"
"

echo ""
echo "Try the following commands to test the deployment externally:"
echo "  curl $SERVICE_URL/health"
echo "  curl -X POST $SERVICE_URL/v1/messages -H \"Content-Type: application/json\" -d '{\"model\":\"claude-3-opus\",\"messages\":[{\"role\":\"user\",\"content\":\"Hello\"}],\"max_tokens\":10}'"
echo "  curl -X POST $SERVICE_URL/generate -H \"Content-Type: application/json\" -d '{\"model\":\"claude\",\"prompt\":\"Hello\",\"options\":{\"temperature\":0.7,\"num_predict\":10},\"stream\":false}'"
echo ""
echo "To run the helm test:"
echo "  helm test ollama-claude-proxy -n $NAMESPACE"
echo ""
echo "To delete the deployment:"
echo "  helm delete ollama-claude-proxy -n $NAMESPACE"