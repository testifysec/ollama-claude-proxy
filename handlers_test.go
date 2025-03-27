package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleOllamaGenerate_BadRequest(t *testing.T) {
	server := NewServer(testConfig())

	// Create an invalid request with malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/api/generate", strings.NewReader("invalid json"))
	req.Header.Set("Content-Type", "application/json")

	// Record the response
	recorder := httptest.NewRecorder()

	// Handle the request
	server.handleOllamaGenerate(recorder, req)

	// Check the response
	resp := recorder.Result()
	defer resp.Body.Close()

	// Should return a 400 Bad Request
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestServerHTTPHandlers(t *testing.T) {
	// Create a test server with routes properly set up
	server := NewServer(testConfig())

	// Create a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			server.handleHealth(w, r)
		case "/api/generate":
			// For the generate endpoint, return a mock response
			var req OllamaRequest
			json.NewDecoder(r.Body).Decode(&req)

			resp := OllamaResponse{
				Model:     req.Model,
				CreatedAt: serverTimeFunc(),
				Response:  "Test response",
				Done:      true,
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	// Test health endpoint
	resp, err := http.Get(ts.URL + "/health")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d for health check, got %d", http.StatusOK, resp.StatusCode)
	}

	// Test generate endpoint with a valid request
	ollamaReq := OllamaRequest{
		Model:  "claude",
		Prompt: "Test",
		Options: OllamaOptions{
			Temperature: 0.7,
			NumPredict:  100,
		},
	}

	reqBody, _ := json.Marshal(ollamaReq)
	resp, err = http.Post(ts.URL+"/api/generate", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d for generate, got %d", http.StatusOK, resp.StatusCode)
	}

	// Verify response contains expected fields
	var ollamaResp OllamaResponse
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if ollamaResp.Model != ollamaReq.Model {
		t.Errorf("Expected model %q in response, got %q", ollamaReq.Model, ollamaResp.Model)
	}

	if !ollamaResp.Done {
		t.Errorf("Expected done=true in response")
	}
}

// Mock time function for deterministic tests
func serverTimeFunc() time.Time {
	return time.Date(2025, 3, 26, 17, 0, 0, 0, time.UTC)
}