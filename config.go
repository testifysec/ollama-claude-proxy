package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	// Server configuration
	Port string `json:"port"`

	// Claude API configuration
	APIKey             string `json:"api_key"`
	APIVersion         string `json:"api_version"`
	APIEndpoint        string `json:"api_endpoint"`
	SystemPrompt       string `json:"system_prompt"`
	DefaultModel       string `json:"default_model"`
	RequestTimeoutSecs int    `json:"request_timeout_secs"`
}

// DefaultConfig returns the default configuration
func DefaultConfig() Config {
	return Config{
		Port:               "8080",
		APIVersion:         "2023-06-01",
		APIEndpoint:        "https://api.anthropic.com/v1/messages",
		SystemPrompt:       "You are Claude, an AI assistant by Anthropic.",
		DefaultModel:       "claude-3-5-sonnet-20240620",
		RequestTimeoutSecs: 60,
	}
}

// LoadConfig loads the configuration from various sources
// Priority order:
// 1. Environment variables (highest priority)
// 2. Config file (if provided)
// 3. Default values (lowest priority)
func LoadConfig(configPath string) (Config, error) {
	// Start with default config
	config := DefaultConfig()

	// Load config file if provided
	if configPath != "" {
		if err := loadConfigFile(&config, configPath); err != nil {
			return config, err
		}
	}

	// Override with environment variables
	applyEnvVars(&config)

	// Validate config
	if err := validateConfig(config); err != nil {
		return config, err
	}

	return config, nil
}

// loadConfigFile loads configuration from a JSON file
func loadConfigFile(config *Config, configPath string) error {
	// Expand path if needed
	expandedPath, err := filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("failed to resolve config path: %w", err)
	}

	// Open file
	file, err := os.Open(expandedPath)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	// Read and parse JSON
	bytes, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	if err := json.Unmarshal(bytes, config); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	log.Printf("Loaded configuration from %s", expandedPath)
	return nil
}

// applyEnvVars overrides config with environment variables
func applyEnvVars(config *Config) {
	// Server config
	if port := os.Getenv("PORT"); port != "" {
		config.Port = port
	}

	// API config
	if apiKey := os.Getenv("ANTHROPIC_API_KEY"); apiKey != "" {
		config.APIKey = apiKey
	}

	if apiVersion := os.Getenv("CLAUDE_API_VERSION"); apiVersion != "" {
		config.APIVersion = apiVersion
	}

	if apiEndpoint := os.Getenv("CLAUDE_API_ENDPOINT"); apiEndpoint != "" {
		config.APIEndpoint = apiEndpoint
	}

	if systemPrompt := os.Getenv("CLAUDE_SYSTEM_PROMPT"); systemPrompt != "" {
		config.SystemPrompt = systemPrompt
	}

	if defaultModel := os.Getenv("CLAUDE_DEFAULT_MODEL"); defaultModel != "" {
		config.DefaultModel = defaultModel
	}

	if timeoutStr := os.Getenv("REQUEST_TIMEOUT_SECS"); timeoutStr != "" {
		var timeout int
		if _, err := fmt.Sscanf(timeoutStr, "%d", &timeout); err == nil && timeout > 0 {
			config.RequestTimeoutSecs = timeout
		}
	}
}

// validateConfig validates the configuration
func validateConfig(config Config) error {
	// APIKey is required
	if config.APIKey == "" {
		return fmt.Errorf("API key is required")
	}

	// Validate timeout is reasonable
	if config.RequestTimeoutSecs <= 0 {
		return fmt.Errorf("request timeout must be positive")
	}

	return nil
}
