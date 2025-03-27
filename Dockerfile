# Build stage
FROM golang:1.23.4-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o ollama-claude-proxy

# Final stage
FROM alpine:3.18

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates

# Create a non-root user and group
RUN addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/ollama-claude-proxy .

# Copy templates directory
COPY templates/ /app/templates/

# Set ownership and permissions
RUN chown -R appuser:appgroup /app && \
    chmod +x /app/ollama-claude-proxy

# Switch to non-root user
USER appuser

# Expose the port the service will run on
EXPOSE 8080

# Set environment variables
ENV PORT=8080

# Run the application
CMD ["/app/ollama-claude-proxy"]