package main

import (
	"testing"

	"github.com/liushuangls/go-anthropic/v2"
)

// Test the buildModelMap function
func TestBuildModelMap(t *testing.T) {
	modelMap := buildModelMap()
	
	// Check that the map is not empty
	if len(modelMap) == 0 {
		t.Fatal("Expected non-empty model map")
	}
	
	// Check some key mappings
	expectedMappings := map[string]anthropic.Model{
		"claude":            anthropic.ModelClaude3Opus20240229,
		"claude-3-opus":     anthropic.ModelClaude3Opus20240229,
		"claude-3-sonnet":   anthropic.ModelClaude3Sonnet20240229,
		"claude-3-haiku":    anthropic.ModelClaude3Haiku20240307,
		"claude-3.5-sonnet": anthropic.ModelClaude3Dot5SonnetLatest,
	}
	
	for name, expectedModel := range expectedMappings {
		if model, exists := modelMap[name]; !exists {
			t.Errorf("Expected mapping for %q", name)
		} else if model != expectedModel {
			t.Errorf("For %q, expected model %q, got %q", name, expectedModel, model)
		}
	}
}

// Test mapModelName with various inputs
func TestMapModelNameComprehensive(t *testing.T) {
	server := NewServer("test-api-key")
	
	testCases := []struct {
		name     string
		input    string
		expected anthropic.Model
	}{
		{"Basic lookup", "claude", anthropic.ModelClaude3Opus20240229},
		{"Case insensitivity", "Claude", anthropic.ModelClaude3Opus20240229},
		{"Mixed case", "Claude-3-Opus", anthropic.ModelClaude3Opus20240229},
		{"All caps", "CLAUDE", anthropic.ModelClaude3Opus20240229},
		{"Specific model", "claude-3-haiku", anthropic.ModelClaude3Haiku20240307},
		{"Unknown model", "unknown-model", defaultClaudeModelID},
		{"Empty string", "", defaultClaudeModelID},
		{"Version format", "claude-3.5-sonnet", anthropic.ModelClaude3Dot5SonnetLatest},
		{"Explicit version", "claude-3.7-sonnet", anthropic.ModelClaude3Dot7SonnetLatest},
	}
	
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := server.mapModelName(tc.input)
			if result != tc.expected {
				t.Errorf("mapModelName(%q) = %q, expected %q", tc.input, result, tc.expected)
			}
		})
	}
}