package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	configDirName  = "gheppo"
	configFileName = "config.json"

	SourceGitHub   = "github"
	SourceLeetCode = "leetcode"
	DefaultSource  = SourceGitHub
)

// Config represents persistent user configuration for Gheppo.
type Config struct {
	Theme            string `json:"theme"`
	Source           string `json:"source,omitempty"`
	LeetCodeUsername string `json:"leetcodeUsername,omitempty"`
}

// ConfigDirectoryFunc allows overriding the config directory in tests.
var configDirectory = defaultConfigDir

func defaultConfigDir() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		// Fallback to home dir
		home, hErr := os.UserHomeDir()
		if hErr != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}

	dir = filepath.Join(dir, configDirName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return "", err
	}
	return dir, nil
}

// ConfigFilePath returns the absolute path to the configuration file.
func ConfigFilePath() (string, error) {
	dir, err := configDirectory()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, configFileName), nil
}

// LoadConfig reads the configuration from disk, returning defaults if not found or corrupted.
func LoadConfig() (*Config, error) {
	path, err := ConfigFilePath()
	if err != nil {
		return &Config{Theme: DefaultThemeName}, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// If file doesn't exist, return default config without error
		return &Config{Theme: DefaultThemeName}, nil
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		// On corrupted JSON, fallback gracefully to default
		return &Config{Theme: DefaultThemeName}, nil
	}

	if cfg.Theme == "" {
		cfg.Theme = DefaultThemeName
	}

	return &cfg, nil
}

// SaveConfig persists configuration to disk.
func SaveConfig(cfg *Config) error {
	if cfg == nil {
		cfg = &Config{Theme: DefaultThemeName}
	}

	if cfg.Theme == "" {
		cfg.Theme = DefaultThemeName
	}

	path, err := ConfigFilePath()
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	// Append trailing newline
	data = append(data, '\n')

	tmpPath := path + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("write temp config file: %w", err)
	}

	if err := os.Rename(tmpPath, path); err != nil {
		_ = os.Remove(tmpPath)
		return fmt.Errorf("atomic rename config file: %w", err)
	}

	return nil
}

// GetSelectedThemeName returns the name of the currently active theme.
// Priority order:
// 1. GHEPPO_THEME environment variable (if set and matches a registered theme)
// 2. Persistent config file (`config.json`)
// 3. Default "github"
func GetSelectedThemeName() string {
	if envTheme := strings.ToLower(strings.TrimSpace(os.Getenv("GHEPPO_THEME"))); envTheme != "" {
		if IsValidTheme(envTheme) {
			return envTheme
		}
	}

	cfg, err := LoadConfig()
	if err == nil && cfg != nil && cfg.Theme != "" {
		if IsValidTheme(cfg.Theme) {
			return strings.ToLower(cfg.Theme)
		}
	}

	return DefaultThemeName
}

// GetSelectedTheme returns the resolved Theme object for the active theme.
func GetSelectedTheme() Theme {
	name := GetSelectedThemeName()
	return ResolveTheme(name)
}

// SetTheme updates and persists the active theme choice.
func SetTheme(name string) error {
	normalized := strings.ToLower(strings.TrimSpace(name))
	if !IsValidTheme(normalized) {
		return fmt.Errorf("unknown theme %q (available: %s)", name, strings.Join(AvailableThemeNames(), ", "))
	}

	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		cfg = &Config{}
	}

	cfg.Theme = normalized
	return SaveConfig(cfg)
}

// AvailableSources returns all supported data source identifiers.
func AvailableSources() []string {
	return []string{SourceGitHub, SourceLeetCode}
}

// IsValidSource reports whether the given source name is supported.
func IsValidSource(name string) bool {
	norm := strings.ToLower(strings.TrimSpace(name))
	return norm == SourceGitHub || norm == SourceLeetCode
}

// IsSourceConfigured reports whether a data source has been explicitly configured.
func IsSourceConfigured() bool {
	if envSrc := strings.ToLower(strings.TrimSpace(os.Getenv("GHEPPO_SOURCE"))); envSrc != "" {
		return IsValidSource(envSrc)
	}

	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		return false
	}
	return cfg.Source != "" && IsValidSource(cfg.Source)
}

// GetSource returns the currently active data source.
// Priority:
// 1. GHEPPO_SOURCE environment variable
// 2. Persistent config file
// 3. Default "github"
func GetSource() string {
	if envSrc := strings.ToLower(strings.TrimSpace(os.Getenv("GHEPPO_SOURCE"))); envSrc != "" {
		if IsValidSource(envSrc) {
			return envSrc
		}
	}

	cfg, err := LoadConfig()
	if err == nil && cfg != nil && cfg.Source != "" {
		if IsValidSource(cfg.Source) {
			return strings.ToLower(cfg.Source)
		}
	}

	return DefaultSource
}

// SetSource updates and persists the active source choice.
func SetSource(source string) error {
	normalized := strings.ToLower(strings.TrimSpace(source))
	if !IsValidSource(normalized) {
		return fmt.Errorf("unknown source %q (available: %s)", source, strings.Join(AvailableSources(), ", "))
	}

	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		cfg = &Config{}
	}

	cfg.Source = normalized
	return SaveConfig(cfg)
}

// GetLeetCodeUsername returns the configured LeetCode username.
func GetLeetCodeUsername() string {
	if envUser := strings.TrimSpace(os.Getenv("GHEPPO_LEETCODE_USERNAME")); envUser != "" {
		return envUser
	}

	cfg, err := LoadConfig()
	if err == nil && cfg != nil {
		return cfg.LeetCodeUsername
	}
	return ""
}

// SetLeetCodeUsername updates and persists the LeetCode username.
func SetLeetCodeUsername(username string) error {
	cfg, err := LoadConfig()
	if err != nil || cfg == nil {
		cfg = &Config{}
	}

	cfg.LeetCodeUsername = strings.TrimSpace(username)
	return SaveConfig(cfg)
}

// SetConfigDirForTesting overrides the configuration directory function for unit tests.
func SetConfigDirForTesting(fn func() (string, error)) func() {
	prev := configDirectory
	configDirectory = fn
	return func() {
		configDirectory = prev
	}
}