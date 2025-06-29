package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// Config represents the application configuration
type Config struct {
	DefaultDirection string   `json:"default_direction"` // "es2en" or "en2es"
	DefaultTenses    []string `json:"default_tenses"`    // Which tenses to show by default
	ShowAllTenses    bool     `json:"show_all_tenses"`   // Show all available tenses
}

// DefaultConfig returns the default configuration
func DefaultConfig() *Config {
	return &Config{
		DefaultDirection: "es2en",
		DefaultTenses:    []string{"present", "preterite"},
		ShowAllTenses:    false,
	}
}

// LoadConfig loads configuration from file or creates default
func LoadConfig() (*Config, error) {
	configPath := getConfigPath()

	// If config file doesn't exist, create default
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		if err := config.Save(); err != nil {
			return config, fmt.Errorf("failed to save default config: %w", err)
		}
		return config, nil
	}

	// Load existing config
	file, err := os.Open(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open config file: %w", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// Save saves the configuration to file
func (c *Config) Save() error {
	configPath := getConfigPath()

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(configPath), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

// getConfigPath returns the path to the configuration file
func getConfigPath() string {
	// Use user's home directory for config
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// Fallback to current directory
		return ".tr-config.json"
	}

	return filepath.Join(homeDir, ".config", "tr", "config.json")
}

// GetAvailableTenses returns all available tenses for conjugation
func GetAvailableTenses() []string {
	return []string{
		"present",
		"preterite",
		"imperfect",
		"future",
		"conditional",
		"present_subjunctive",
		"imperfect_subjunctive",
		"present_perfect",
		"pluperfect",
		"future_perfect",
		"conditional_perfect",
		"present_perfect_subjunctive",
	}
}
