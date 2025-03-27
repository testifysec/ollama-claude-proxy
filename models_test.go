package main

import (
	"testing"
)

// Test the buildModelMap function
func TestBuildModelMap(t *testing.T) {
	modelMap := buildModelMap()

	// Check that the map is not empty
	if len(modelMap) == 0 {
		t.Fatal("Expected non-empty model map")
	}

	// Check some key mappings
	expectedMappings := map[string]ModelID{
		"claude":            ModelClaude3Opus20240229,
		"claude-3-opus":     ModelClaude3Opus20240229,
		"claude-3-sonnet":   ModelClaude3Sonnet20240229,
		"claude-3-haiku":    ModelClaude3Haiku20240307,
		"claude-3.5-sonnet": ModelClaude3Dot5SonnetLatest,
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
		expected ModelID
	}{
		{"Basic lookup", "claude", ModelClaude3Opus20240229},
		{"Case insensitivity", "Claude", ModelClaude3Opus20240229},
		{"Mixed case", "Claude-3-Opus", ModelClaude3Opus20240229},
		{"All caps", "CLAUDE", ModelClaude3Opus20240229},
		{"Specific model", "claude-3-haiku", ModelClaude3Haiku20240307},
		{"Unknown model", "unknown-model", defaultClaudeModelID},
		{"Empty string", "", defaultClaudeModelID},
		{"Version format", "claude-3.5-sonnet", ModelClaude3Dot5SonnetLatest},
		{"Explicit version", "claude-3.7-sonnet", ModelClaude3Dot7SonnetLatest},
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
