package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// ModelID represents a Claude model identifier
type ModelID string

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

// Claude API request structures
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
)

type MessageContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type Message struct {
	Role    MessageRole      `json:"role"`
	Content []MessageContent `json:"content"`
}

type ClaudeRequest struct {
	Model     ModelID   `json:"model"`
	Messages  []Message `json:"messages"`
	System    string    `json:"system,omitempty"`
	MaxTokens int       `json:"max_tokens,omitempty"`
	// Optional parameters
	Temperature *float32 `json:"temperature,omitempty"`
	TopP        *float32 `json:"top_p,omitempty"`
	TopK        *int     `json:"top_k,omitempty"`
}

type ClaudeContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type ClaudeResponse struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Role       string          `json:"role"`
	Content    []ClaudeContent `json:"content"`
	StopReason string          `json:"stop_reason"`
}

// Server holds the server configuration and dependencies
type Server struct {
	config    Config
	modelMap  map[string]ModelID
	templates *template.Template
}

// NewServer creates a new proxy server instance
func NewServer(config Config) *Server {
	// Try to parse templates from multiple possible locations
	var tmpl *template.Template
	var err error

	// Try different template paths that might be used in different environments
	templatePaths := []string{
		filepath.Join("templates", "*.html"),
		filepath.Join("/app/templates", "*.html"),
		filepath.Join(".", "templates", "*.html"),
	}

	for _, path := range templatePaths {
		tmpl, err = template.ParseGlob(path)
		if err == nil {
			log.Printf("Successfully loaded templates from: %s", path)
			break
		}
	}

	if err != nil {
		log.Printf("Warning: Failed to parse templates: %v", err)
	}

	return &Server{
		config:    config,
		modelMap:  buildModelMap(config),
		templates: tmpl,
	}
}

// Builds a map of Ollama model names to Claude model IDs
func buildModelMap(config Config) map[string]ModelID {
	return map[string]ModelID{
		"claude":            "claude-3-opus-20240229",
		"claude-3":          "claude-3-opus-20240229",
		"claude-3-opus":     "claude-3-opus-20240229",
		"claude-3-sonnet":   "claude-3-sonnet-20240229",
		"claude-3-haiku":    "claude-3-haiku-20240307",
		"claude-2":          "claude-2.0",
		"claude-2.1":        "claude-2.1",
		"claude-3.5":        "claude-3-5-sonnet-20240620",
		"claude-3.5-sonnet": "claude-3-5-sonnet-20240620",
		"claude-3.7":        "claude-3-7-sonnet-20240610",
		"claude-3.7-sonnet": "claude-3-7-sonnet-20240610",
	}
}

// Map Ollama model names to Claude model IDs
func (s *Server) mapModelName(name string) ModelID {
	// Convert to lowercase for case-insensitive matching
	name = strings.ToLower(name)

	if model, exists := s.modelMap[name]; exists {
		return model
	}

	// Default to default model if not found
	return ModelID(s.config.DefaultModel)
}

// Create a user message from text
func NewUserTextMessage(text string) Message {
	return Message{
		Role: RoleUser,
		Content: []MessageContent{
			{
				Type: "text",
				Text: text,
			},
		},
	}
}

// Call the Claude API directly
func (s *Server) callClaudeAPI(ctx context.Context, claudeReq ClaudeRequest) (*ClaudeResponse, error) {
	// Get API key from environment if not in config
	apiKey := s.config.APIKey
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("API key not found in config or environment")
		}
	}

	// Marshal the request body
	reqBody, err := json.Marshal(claudeReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create the HTTP request
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.config.APIEndpoint, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Anthropic-Version", s.config.APIVersion)

	// Send the request
	client := &http.Client{Timeout: time.Duration(s.config.RequestTimeoutSecs) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to call Claude API: %w", err)
	}
	defer resp.Body.Close()

	// Check for error status code
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("claude API returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var claudeResp ClaudeResponse
	if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &claudeResp, nil
}

// Extract the first text from a Claude response
func getFirstContentText(resp *ClaudeResponse) string {
	if resp == nil || len(resp.Content) == 0 {
		return ""
	}

	for _, content := range resp.Content {
		if content.Type == "text" {
			return content.Text
		}
	}

	return ""
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
	claudeReq := ClaudeRequest{
		Model: claudeModel,
		Messages: []Message{
			NewUserTextMessage(ollamaReq.Prompt),
		},
		System:    s.config.SystemPrompt,
		MaxTokens: ollamaReq.Options.NumPredict,
	}

	// Set optional parameters
	if ollamaReq.Options.Temperature > 0 {
		temp := float32(ollamaReq.Options.Temperature)
		claudeReq.Temperature = &temp
	}

	if ollamaReq.Options.TopP > 0 {
		topP := float32(ollamaReq.Options.TopP)
		claudeReq.TopP = &topP
	}

	if ollamaReq.Options.TopK > 0 {
		claudeReq.TopK = &ollamaReq.Options.TopK
	}

	// Log request for debugging
	if reqJSON, err := json.MarshalIndent(claudeReq, "", "  "); err == nil {
		log.Printf("Claude API request (via Ollama compat): %s", string(reqJSON))
	}

	// Send the request to Claude
	resp, err := s.callClaudeAPI(context.Background(), claudeReq)
	if err != nil {
		log.Printf("Error calling Claude API: %v", err)
		http.Error(w, fmt.Sprintf("Claude API error: %v", err), http.StatusBadGateway)
		return
	}

	// Extract text from response
	responseText := getFirstContentText(resp)

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

// Serve the UI
func (s *Server) handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	if s.templates == nil {
		http.Error(w, "UI templates not available", http.StatusInternalServerError)
		return
	}

	// Get models for display
	models := make([]struct {
		Name  string
		Value string
	}, 0, len(s.modelMap))

	for name, id := range s.modelMap {
		models = append(models, struct {
			Name  string
			Value string
		}{
			Name:  name + " (" + string(id) + ")",
			Value: name,
		})
	}

	templateData := struct {
		SystemPrompt   string
		DefaultModel   string
		APIEndpoint    string
		RequestTimeout int
		Models         []struct {
			Name  string
			Value string
		}
	}{
		SystemPrompt:   s.config.SystemPrompt,
		DefaultModel:   s.config.DefaultModel,
		APIEndpoint:    s.config.APIEndpoint,
		RequestTimeout: s.config.RequestTimeoutSecs,
		Models:         models,
	}

	if err := s.templates.ExecuteTemplate(w, "index.html", templateData); err != nil {
		log.Printf("Error rendering template: %v", err)
		http.Error(w, "Error rendering UI", http.StatusInternalServerError)
		return
	}
}

// Setup routes and start the server
func (s *Server) Start(port string) error {
	// Setup routes
	http.HandleFunc("/health", s.handleHealth)
	http.HandleFunc("/", s.handleUI)

	// Setup API routes with CORS
	http.HandleFunc("/api/generate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		s.handleOllamaGenerate(w, r)
	})

	// Start the server
	log.Printf("Ollama-Claude proxy listening on port %s...", port)
	log.Printf("UI available at http://localhost:%s/", port)
	return http.ListenAndServe(":"+port, nil)
}

func main() {
	// Define command-line flags
	configPathPtr := flag.String("config", "", "Path to configuration file")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig(*configPathPtr)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create and start server
	server := NewServer(config)
	log.Fatal(server.Start(config.Port))
}
