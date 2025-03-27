package main

import (
	"testing"
)

// Test the buildModelMap function
func TestBuildModelMap(t *testing.T) {
	config := testConfig()
	modelMap := buildModelMap(config)

	// Check that the map is not empty
	if len(modelMap) == 0 {
		t.Fatal("Expected non-empty model map")
	}

	// Check some key mappings
	expectedMappings := map[string]ModelID{
		"claude":            testModelOpus,
		"claude-3-opus":     testModelOpus,
		"claude-3-sonnet":   testModelSonnet,
		"claude-3-haiku":    testModelHaiku,
		"claude-3.5-sonnet": testModelSonnet35,
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
	server := NewServer(testConfig())

	testCases := []struct {
		name     string
		input    string
		expected ModelID
	}{
		{"Basic lookup", "claude", testModelOpus},
		{"Case insensitivity", "Claude", testModelOpus},
		{"Mixed case", "Claude-3-Opus", testModelOpus},
		{"All caps", "CLAUDE", testModelOpus},
		{"Specific model", "claude-3-haiku", testModelHaiku},
		{"Unknown model", "unknown-model", testModelSonnet35},
		{"Empty string", "", testModelSonnet35},
		{"Version format", "claude-3.5-sonnet", testModelSonnet35},
		{"Explicit version", "claude-3.7-sonnet", ModelID("claude-3-7-sonnet-20240610")},
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