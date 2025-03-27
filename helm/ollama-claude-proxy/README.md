# Ollama-Claude Proxy Helm Chart

This Helm chart deploys the Ollama-Claude proxy, which allows you to use Anthropic's Claude API with Ollama-compatible clients.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.2.0+

## Getting Started

### Add the Helm Repository

```bash
# If hosting in a Helm repository
helm repo add my-repo https://charts.example.com/
helm repo update
```

### Installing the Chart

To install the chart with the release name `my-release`:

```bash
# Using a stored Anthropic API key
helm install my-release my-repo/ollama-claude-proxy \
  --set env.ANTHROPIC_API_KEY=sk-ant-your-api-key-here

# Using an existing secret
helm install my-release my-repo/ollama-claude-proxy \
  --set secret.existingSecret=my-secrets \
  --set secret.existingSecretKey=anthropic-api-key
```

The command deploys the Ollama-Claude proxy on the Kubernetes cluster in the default configuration.

### Uninstalling the Chart

To uninstall/delete the `my-release` deployment:

```bash
helm delete my-release
```

## Configuration

The following table lists the configurable parameters of the Ollama-Claude proxy chart and their default values.

| Parameter                               | Description                                                                   | Default                       |
|-----------------------------------------|-------------------------------------------------------------------------------|-------------------------------|
| `replicaCount`                          | Number of replicas                                                            | `1`                           |
| `image.repository`                      | Image repository                                                              | `ollama-claude-proxy`         |
| `image.tag`                             | Image tag                                                                     | `latest`                      |
| `image.pullPolicy`                      | Image pull policy                                                             | `IfNotPresent`                |
| `service.type`                          | Kubernetes Service type                                                       | `ClusterIP`                   |
| `service.port`                          | Service port                                                                  | `8080`                        |
| `resources.limits.cpu`                  | CPU resource limits                                                           | `200m`                        |
| `resources.limits.memory`               | Memory resource limits                                                        | `128Mi`                       |
| `resources.requests.cpu`                | CPU resource requests                                                         | `100m`                        |
| `resources.requests.memory`             | Memory resource requests                                                      | `64Mi`                        |
| `env.ANTHROPIC_API_KEY`                 | Anthropic API key (only if not using existingSecret)                          | `""`                          |
| `env.PORT`                              | Port to run the service on                                                    | `"8080"`                      |
| `secret.create`                         | Whether to create a Secret                                                    | `true`                        |
| `secret.name`                           | Name of the Secret                                                            | `ollama-claude-proxy`         |
| `secret.key`                            | Key in the Secret for the API key                                             | `anthropic-api-key`           |
| `secret.existingSecret`                 | Name of an existing Secret                                                    | `""`                          |
| `secret.existingSecretKey`              | Key in the existing Secret                                                    | `""`                          |
| `probes.liveness.enabled`               | Enable liveness probe                                                         | `true`                        |
| `probes.liveness.initialDelaySeconds`   | Initial delay for liveness probe                                              | `10`                          |
| `probes.liveness.periodSeconds`         | Period for liveness probe                                                     | `10`                          |
| `probes.liveness.timeoutSeconds`        | Timeout for liveness probe                                                    | `5`                           |
| `probes.liveness.failureThreshold`      | Failure threshold for liveness probe                                          | `3`                           |
| `probes.readiness.enabled`              | Enable readiness probe                                                        | `true`                        |
| `probes.readiness.initialDelaySeconds`  | Initial delay for readiness probe                                             | `5`                           |
| `probes.readiness.periodSeconds`        | Period for readiness probe                                                    | `10`                          |
| `probes.readiness.timeoutSeconds`       | Timeout for readiness probe                                                   | `5`                           |
| `probes.readiness.failureThreshold`     | Failure threshold for readiness probe                                         | `3`                           |
| `ingress.enabled`                       | Enable ingress                                                                | `false`                       |
| `ingress.className`                     | Ingress class name                                                            | `""`                          |
| `ingress.annotations`                   | Ingress annotations                                                           | `{}`                          |
| `ingress.hosts`                         | Ingress hosts                                                                 | See `values.yaml`             |
| `ingress.tls`                           | Ingress TLS configuration                                                     | `[]`                          |

## Example: Configuration with Values File

Create a `values.yaml` file:

```yaml
replicaCount: 2

image:
  repository: myrepo/ollama-claude-proxy
  tag: 1.0.0

service:
  type: LoadBalancer
  port: 8080

secret:
  existingSecret: "api-keys"
  existingSecretKey: "anthropic-api-key"

resources:
  limits:
    cpu: 500m
    memory: 256Mi
  requests:
    cpu: 200m
    memory: 128Mi

ingress:
  enabled: true
  className: "nginx"
  annotations:
    kubernetes.io/ingress.class: nginx
    cert-manager.io/cluster-issuer: letsencrypt-prod
  hosts:
    - host: ollama-claude-proxy.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: ollama-claude-proxy-tls
      hosts:
        - ollama-claude-proxy.example.com
```

Then install the chart:

```bash
helm install my-release my-repo/ollama-claude-proxy -f values.yaml
```