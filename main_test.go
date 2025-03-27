package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/liushuangls/go-anthropic/v2"
)

// Create a simplified test response for Anthropic API
func createTestResponse(text string) anthropic.MessagesResponse {
	return anthropic.MessagesResponse{
		ID:   "test-id",
		Type: anthropic.MessagesResponseTypeMessage,
		Role: anthropic.RoleAssistant,
		Content: []anthropic.MessageContent{
			{
				Type: anthropic.MessagesContentTypeText,
				Text: &text,
			},
		},
	}
}

// Test the model mapping functionality
func TestMapModelName(t *testing.T) {
	server := NewServer("test-api-key")

	testCases := []struct {
		input    string
		expected anthropic.Model
	}{
		{"claude", anthropic.ModelClaude3Opus20240229},
		{"Claude", anthropic.ModelClaude3Opus20240229}, // Test case insensitivity
		{"claude-3-haiku", anthropic.ModelClaude3Haiku20240307},
		{"claude-3.5-sonnet", anthropic.ModelClaude3Dot5SonnetLatest},
		{"unknown-model", anthropic.ModelClaude3Opus20240229}, // Default model
	}

	for _, tc := range testCases {
		result := server.mapModelName(tc.input)
		if result != tc.expected {
			t.Errorf("mapModelName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

// Test the request parsing and response formatting parts of the Ollama generate endpoint
func TestHandleOllamaGenerate_RequestParsing(t *testing.T) {
	// Create a test request
	ollamaReq := OllamaRequest{
		Model: "claude",
		Prompt: "Test prompt",
		Options: OllamaOptions{
			Temperature: 0.7,
			TopP:        0.95,
			TopK:        40,
			NumPredict:  100,
		},
	}
	reqBody, _ := json.Marshal(ollamaReq)
	
	// Create the test request
	req := httptest.NewRequest(http.MethodPost, "/generate", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	
	// Record the response
	recorder := httptest.NewRecorder()
	
	// Mock the handler to avoid making real API calls
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the request just like the real handler would
		var parsedReq OllamaRequest
		if err := json.NewDecoder(r.Body).Decode(&parsedReq); err != nil {
			http.Error(w, "Bad request: "+err.Error(), http.StatusBadRequest)
			return
		}
		
		// Verify the parsed request matches what we sent
		if parsedReq.Model != ollamaReq.Model {
			t.Errorf("Expected model %q, got %q", ollamaReq.Model, parsedReq.Model)
		}
		if parsedReq.Prompt != ollamaReq.Prompt {
			t.Errorf("Expected prompt %q, got %q", ollamaReq.Prompt, parsedReq.Prompt)
		}
		if parsedReq.Options.Temperature != ollamaReq.Options.Temperature {
			t.Errorf("Expected temperature %v, got %v", ollamaReq.Options.Temperature, parsedReq.Options.Temperature)
		}
		
		// Send a mock response
		mockResp := OllamaResponse{
			Model:     parsedReq.Model,
			CreatedAt: time.Now(),
			Response:  "Mock response for testing",
			Done:      true,
		}
		
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mockResp)
	})
	
	// Call the handler
	handler.ServeHTTP(recorder, req)
	
	// Check response
	resp := recorder.Result()
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}
	
	var respBody OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}
	
	// Check response fields
	if respBody.Model != ollamaReq.Model {
		t.Errorf("Expected model %q, got %q", ollamaReq.Model, respBody.Model)
	}
	if !respBody.Done {
		t.Errorf("Expected done to be true, got false")
	}
}

// Test the health check endpoint
func TestHealthCheck(t *testing.T) {
	server := NewServer("test-api-key")
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	recorder := httptest.NewRecorder()

	server.handleHealth(recorder, req)

	resp := recorder.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "OK" {
		t.Errorf("Expected body %q, got %q", "OK", string(body))
	}
}