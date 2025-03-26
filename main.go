package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

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

type ClaudeRequest struct {
	Model             string   `json:"model"`
	Prompt            string   `json:"prompt"`
	MaxTokensToSample int      `json:"max_tokens_to_sample"`
	Temperature       float64  `json:"temperature,omitempty"`
	TopP              float64  `json:"top_p,omitempty"`
	TopK              int      `json:"top_k,omitempty"`
	Stream            bool     `json:"stream,omitempty"`
	StopSequences     []string `json:"stop_sequences,omitempty"`
}

type ClaudeResponse struct {
	Completion string `json:"completion"`
	StopReason string `json:"stop_reason"`
}

type ProxyResponse struct {
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	Response  string    `json:"response"`
	Done      bool      `json:"done"`
}

func mapModelName(name string) string {
	// Map simple names to Claude model IDs
	modelMap := map[string]string{
		"claude":          "claude-3-5-sonnet-20241022",
		"claude-3-sonnet": "claude-3-sonnet-20240229",
		"claude-3-opus":   "claude-3-opus-20240229",
		"claude-3-haiku":  "claude-3-haiku-20240307",
	}

	if model, exists := modelMap[name]; exists {
		return model
	}

	// If not in our map, return as-is (might be a full model name already)
	return name
}

func formatPrompt(userPrompt string) string {
	// Wrap the prompt for completions API
	return "\n\nHuman: " + userPrompt + "\n\nAssistant:"
}

func main() {
	// Get API key from environment variables
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		log.Fatal("Anthropic API key not set. Please set ANTHROPIC_API_KEY environment variable.")
	}

	http.HandleFunc("/generate", func(w http.ResponseWriter, r *http.Request) {
		// Parse Ollama-style request
		var ollamaReq OllamaRequest
		if err := json.NewDecoder(r.Body).Decode(&ollamaReq); err != nil {
			http.Error(w, "Bad Request: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Prepare Claude API request
		claudeReq := ClaudeRequest{
			Model:             mapModelName(ollamaReq.Model),
			Prompt:            formatPrompt(ollamaReq.Prompt),
			MaxTokensToSample: ollamaReq.Options.NumPredict,
			Temperature:       ollamaReq.Options.Temperature,
			TopP:              ollamaReq.Options.TopP,
			TopK:              ollamaReq.Options.TopK,
			Stream:            false,
		}
		reqBody, _ := json.Marshal(claudeReq)

		// Send request to Anthropic Claude API
		apiURL := "https://api.anthropic.com/v1/complete"
		req, _ := http.NewRequest("POST", apiURL, bytes.NewReader(reqBody))
		req.Header.Set("x-api-key", apiKey)
		req.Header.Set("anthropic-version", "2023-06-01")
		req.Header.Set("Content-Type", "application/json")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			http.Error(w, "Claude API error: "+err.Error(), http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		// Check response status
		if resp.StatusCode != http.StatusOK {
			// Read error response
			body, _ := io.ReadAll(resp.Body)
			http.Error(w, "Claude API error: "+string(body), resp.StatusCode)
			return
		}

		// Read Claude API response
		var claudeResp ClaudeResponse
		if err := json.NewDecoder(resp.Body).Decode(&claudeResp); err != nil {
			http.Error(w, "Failed to parse Claude response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Construct proxy response in Ollama format
		proxyResp := ProxyResponse{
			Model:     ollamaReq.Model,
			CreatedAt: time.Now(),
			Response:  claudeResp.Completion,
			Done:      true,
		}
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(proxyResp); err != nil {
			http.Error(w, "Failed to send response: "+err.Error(), http.StatusInternalServerError)
		}
	})

	// Health check endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Ollama-Claude proxy listening on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
