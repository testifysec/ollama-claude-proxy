package main

import (
	"testing"
)

// Test server initialization and configuration
func TestNewServer(t *testing.T) {
	// Test with a valid API key
	config := testConfig()
	server := NewServer(config)
	
	// Check that the server has been properly initialized
	if server.config.APIKey != config.APIKey {
		t.Errorf("Expected API key to be %q, got %q", config.APIKey, server.config.APIKey)
	}
	
	if len(server.modelMap) == 0 {
		t.Error("Expected non-empty model map")
	}
	
	// Check model mappings
	if modelID := server.mapModelName("claude"); modelID != testModelOpus {
		t.Errorf("Expected model ID for 'claude' to be %q, got %q", testModelOpus, modelID)
	}
}