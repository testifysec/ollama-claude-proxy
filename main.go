package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/liushuangls/go-anthropic/v2"
)

// Configuration constants
const (
	defaultPort            = "8080"
	claudeAPIVersion       = "2023-06-01"
	claudeAPIURL           = "https://api.anthropic.com/v1/messages"
	defaultSystemPrompt    = "You are Claude, an AI assistant by Anthropic."
	defaultClaudeModelID   = anthropic.ModelClaude3Opus20240229
)

// Ollama API structures
type OllamaRequest struct {
	Model   string        `json:"model"`
	Prompt  string        `json:"prompt"`
	Options OllamaOptions `json:"options"`
	Stream  bool          `json:"stream"`
}

type OllamaOptions struct {
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	TopK        int     `json:"top_k"`
	NumPredict  int     `json:"num_predict"`
}

type OllamaResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
}

// Server holds the server configuration and dependencies
type Server struct {
	claudeClient *anthropic.Client
	apiKey       string
	modelMap     map[string]anthropic.Model
}

// NewServer creates a new proxy server instance
func NewServer(apiKey string) *Server {
	return &Server{
		claudeClient: anthropic.NewClient(apiKey),
		apiKey:       apiKey,
		modelMap:     buildModelMap(),
	}
}

// Builds a map of Ollama model names to Claude model IDs
func buildModelMap() map[string]anthropic.Model {
	return map[string]anthropic.Model{
		"claude":            anthropic.ModelClaude3Opus20240229,
		"claude-3":          anthropic.ModelClaude3Opus20240229,
		"claude-3-opus":     anthropic.ModelClaude3Opus20240229,
		"claude-3-sonnet":   anthropic.ModelClaude3Sonnet20240229,
		"claude-3-haiku":    anthropic.ModelClaude3Haiku20240307,
		"claude-2":          anthropic.ModelClaude2Dot0,
		"claude-2.1":        anthropic.ModelClaude2Dot1,
		"claude-3.5":        anthropic.ModelClaude3Dot5SonnetLatest,
		"claude-3.5-sonnet": anthropic.ModelClaude3Dot5SonnetLatest,
		"claude-3.7":        anthropic.ModelClaude3Dot7SonnetLatest,
		"claude-3.7-sonnet": anthropic.ModelClaude3Dot7SonnetLatest,
	}
}

// Map Ollama model names to Claude model IDs
func (s *Server) mapModelName(name string) anthropic.Model {
	// Convert to lowercase for case-insensitive matching
	name = strings.ToLower(name)

	if model, exists := s.modelMap[name]; exists {
		return model
	}
	
	// Default to Claude 3 Opus if model not found
	return defaultClaudeModelID
}

// Handle direct Claude API requests
func (s *Server) handleClaudeMessages(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request to /v1/messages endpoint")

	// Read the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	// Log formatted request (for debugging)
	var prettyJSON map[string]interface{}
	if err := json.Unmarshal(body, &prettyJSON); err == nil {
		if prettyBody, err := json.MarshalIndent(prettyJSON, "", "  "); err == nil {
			log.Printf("Claude Messages API request: %s", string(prettyBody))
		}
	}

	// Forward directly to Anthropic API
	req, err := http.NewRequest("POST", claudeAPIURL, strings.NewReader(string(body)))
	if err != nil {
		http.Error(w, "Error creating request: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", s.apiKey)
	req.Header.Set("anthropic-version", claudeAPIVersion)

	// Forward request
	httpClient := &http.Client{}
	resp, err := httpClient.Do(req)
	if err != nil {
		http.Error(w, "Error forwarding to Claude API: "+err.Error(), http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Read response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Error reading Claude API response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Log formatted response (for debugging)
	var prettyResp map[string]interface{}
	if err := json.Unmarshal(respBody, &prettyResp); err == nil {
		if prettyRespBody, err := json.MarshalIndent(prettyResp, "", "  "); err == nil {
			log.Printf("Claude API response: %s", string(prettyRespBody))
		}
	}

	// Return response with same status code and headers
	for key, values := range resp.Header {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}
	w.WriteHeader(resp.StatusCode)
	w.Write(respBody)
}

// Handle Ollama-compatible requests
func (s *Server) handleOllamaGenerate(w http.ResponseWriter, r *http.Request) {
	// Parse the Ollama request
	var ollamaReq OllamaRequest
	if err := json.NewDecoder(r.Body).Decode(&ollamaReq); err != nil {
		http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
		return
	}

	// Map Ollama model to Claude model
	claudeModel := s.mapModelName(ollamaReq.Model)
	log.Printf("Mapped Ollama model '%s' to Claude model '%s'", ollamaReq.Model, claudeModel)

	// Create the Claude message request
	req := anthropic.MessagesRequest{
		Model: claudeModel,
		Messages: []anthropic.Message{
			anthropic.NewUserTextMessage(ollamaReq.Prompt),
		},
		MaxTokens: ollamaReq.Options.NumPredict,
		System:    defaultSystemPrompt,
	}

	// Set optional parameters
	if ollamaReq.Options.Temperature > 0 {
		req.SetTemperature(float32(ollamaReq.Options.Temperature))
	}
	
	if ollamaReq.Options.TopP > 0 {
		req.SetTopP(float32(ollamaReq.Options.TopP))
	}
	
	if ollamaReq.Options.TopK > 0 {
		req.SetTopK(ollamaReq.Options.TopK)
	}

	// Log request for debugging
	if reqJSON, err := json.MarshalIndent(req, "", "  "); err == nil {
		log.Printf("Claude API request (via Ollama compat): %s", string(reqJSON))
	}

	// Send the request to Claude
	resp, err := s.claudeClient.CreateMessages(context.Background(), req)
	if err != nil {
		log.Printf("Error calling Claude API: %v", err)
		http.Error(w, fmt.Sprintf("Claude API error: %v", err), http.StatusBadGateway)
		return
	}

	// Extract text from response
	responseText := resp.GetFirstContentText()

	// Create Ollama response
	ollamaResp := OllamaResponse{
		Model:     ollamaReq.Model,
		CreatedAt: time.Now(),
		Response:  responseText,
		Done:      true,
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ollamaResp)
}

// Health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

// Setup routes and start the server
func (s *Server) Start(port string) error {
	// Setup routes
	http.HandleFunc("/v1/messages", s.handleClaudeMessages)
	http.HandleFunc("/generate", s.handleOllamaGenerate)
	http.HandleFunc("/health", s.handleHealth)

	// Start the server
	log.Printf("Ollama-Claude proxy listening on port %s...", port)
	return http.ListenAndServe(":"+port, nil)
}

func main() {
	// Get the Anthropic API key from environment
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("ANTHROPIC_API_KEY environment variable not set")
	}

	// Get port from environment or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	// Create and start server
	server := NewServer(apiKey)
	log.Fatal(server.Start(port))
}