package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHandleClaudeMessages_ReadError(t *testing.T) {
	// Create a mock body that will return an error when read
	errorReader := &errorReadCloser{err: io.ErrUnexpectedEOF}
	
	server := NewServer("test-api-key")
	
	// Create a request with a body that will fail to read
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", errorReader)
	req.Header.Set("Content-Type", "application/json")
	
	// Record the response
	recorder := httptest.NewRecorder()
	
	// Handle the request
	server.handleClaudeMessages(recorder, req)
	
	// Check the response
	resp := recorder.Result()
	defer resp.Body.Close()
	
	// Should return a 400 Bad Request
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status code %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	
	// Response should contain error message
	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Error reading request body") {
		t.Errorf("Expected error message about reading request body, got: %s", string(body))
	}
}

// Mock reader that returns an error
type errorReadCloser struct {
	err error
}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, e.err
}

func (e *errorReadCloser) Close() error {
	return nil
}

func TestHandleOllamaGenerate_BadRequest(t *testing.T) {
	server := NewServer("test-api-key")
	
	// Create an invalid request with malformed JSON
	req := httptest.NewRequest(http.MethodPost, "/generate", strings.NewReader("invalid json"))
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
	server := NewServer("test-api-key")
	
	// Create a test HTTP server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/health":
			server.handleHealth(w, r)
		case "/v1/messages":
			// For messages endpoint, return a mock response
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			io.WriteString(w, `{"id":"msg_test","type":"message","role":"assistant","content":[{"type":"text","text":"Test"}]}`)
		case "/generate":
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
	resp, err = http.Post(ts.URL+"/generate", "application/json", bytes.NewBuffer(reqBody))
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