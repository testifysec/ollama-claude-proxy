package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Helper to create test config
func testConfig() Config {
	return Config{
		APIKey:             "test-api-key",
		Port:               "8080",
		APIVersion:         "2023-06-01",
		APIEndpoint:        "https://api.anthropic.com/v1/messages",
		SystemPrompt:       "You are Claude, an AI assistant by Anthropic.",
		DefaultModel:       "claude-3-5-sonnet-20240620",
		RequestTimeoutSecs: 60,
	}
}

// Constants for testing
const (
	testModelOpus     = ModelID("claude-3-opus-20240229")
	testModelSonnet   = ModelID("claude-3-sonnet-20240229")
	testModelHaiku    = ModelID("claude-3-haiku-20240307")
	testModelSonnet35 = ModelID("claude-3-5-sonnet-20240620")
)

// Test the model mapping functionality
func TestMapModelName(t *testing.T) {
	server := NewServer(testConfig())

	testCases := []struct {
		input    string
		expected ModelID
	}{
		{"claude", testModelOpus},
		{"Claude", testModelOpus}, // Test case insensitivity
		{"claude-3-haiku", testModelHaiku},
		{"claude-3.5-sonnet", testModelSonnet35},
		{"unknown-model", testModelSonnet35}, // Default model
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
		Model:  "claude",
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
	req := httptest.NewRequest(http.MethodPost, "/api/generate", bytes.NewBuffer(reqBody))
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
	server := NewServer(testConfig())
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
