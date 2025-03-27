package main

import (
	"testing"
)

// Test server initialization and configuration
func TestNewServer(t *testing.T) {
	// Test with a valid API key
	server := NewServer("test-api-key")
	
	// Check that the server has been properly initialized
	if server.apiKey != "test-api-key" {
		t.Errorf("Expected API key to be %q, got %q", "test-api-key", server.apiKey)
	}
	
	if server.claudeClient == nil {
		t.Error("Expected non-nil Claude client")
	}
	
	if len(server.modelMap) == 0 {
		t.Error("Expected non-empty model map")
	}
	
	// Check model mappings
	if modelID := server.mapModelName("claude"); modelID != defaultClaudeModelID {
		t.Errorf("Expected default model ID for 'claude', got %q", modelID)
	}
}