package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestThemeRegistry(t *testing.T) {
	expectedThemes := []string{"github", "mono", "catppuccin", "nord", "gruvbox"}

	// Verify all expected themes are registered
	names := AvailableThemeNames()
	if len(names) != len(expectedThemes) {
		t.Fatalf("expected %d themes, got %d", len(expectedThemes), len(names))
	}

	for _, expected := range expectedThemes {
		theme, ok := GetTheme(expected)
		if !ok {
			t.Errorf("expected theme %q to be registered", expected)
		}
		if theme.Name != expected {
			t.Errorf("theme.Name = %q, want %q", theme.Name, expected)
		}
		if theme.Accent != theme.GraphLevels[4] {
			t.Errorf("theme %q Accent = %q, want GraphLevels[4] = %q", expected, theme.Accent, theme.GraphLevels[4])
		}
		if theme.GraphEmpty != theme.GraphLevels[0] {
			t.Errorf("theme %q GraphEmpty = %q, want GraphLevels[0] = %q", expected, theme.GraphEmpty, theme.GraphLevels[0])
		}
		if theme.Foreground == "" || theme.Muted == "" || theme.Background == "" {
			t.Errorf("theme %q has missing token definitions", expected)
		}
	}

	// Verify case-insensitivity
	if _, ok := GetTheme("Catppuccin"); !ok {
		t.Errorf("expected case-insensitive lookup to succeed for Catppuccin")
	}
	if _, ok := GetTheme("GITHUB"); !ok {
		t.Errorf("expected case-insensitive lookup to succeed for GITHUB")
	}

	// Verify invalid theme lookup
	if _, ok := GetTheme("nonexistent_theme"); ok {
		t.Errorf("expected nonexistent theme lookup to return false")
	}

	// Verify ResolveTheme fallback
	fallback := ResolveTheme("nonexistent_theme")
	if fallback.Name != DefaultThemeName {
		t.Errorf("ResolveTheme(nonexistent) = %q, want %q", fallback.Name, DefaultThemeName)
	}

	defaultTheme := DefaultTheme()
	if defaultTheme.Name != "github" {
		t.Errorf("DefaultTheme() = %q, want github", defaultTheme.Name)
	}
}

func TestConfigLoadAndSave(t *testing.T) {
	tempDir := t.TempDir()
	restore := SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	// 1. Loading with missing file returns default config without error
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("unexpected error loading missing config: %v", err)
	}
	if cfg.Theme != DefaultThemeName {
		t.Errorf("LoadConfig() on missing file returned Theme = %q, want %q", cfg.Theme, DefaultThemeName)
	}

	// 2. Set theme to "nord" and save
	if err := SetTheme("nord"); err != nil {
		t.Fatalf("SetTheme(nord) failed: %v", err)
	}

	// 3. Verify saved file exists and contains "nord"
	filePath := filepath.Join(tempDir, "config.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatalf("reading config file failed: %v", err)
	}
	if !containsSubstring(string(data), `"theme": "nord"`) {
		t.Errorf("config file content = %s, want theme nord", string(data))
	}

	// 4. Verify LoadConfig retrieves the saved theme
	loaded, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() failed: %v", err)
	}
	if loaded.Theme != "nord" {
		t.Errorf("loaded.Theme = %q, want nord", loaded.Theme)
	}

	// 5. Verify GetSelectedTheme returns nord
	selected := GetSelectedTheme()
	if selected.Name != "nord" {
		t.Errorf("GetSelectedTheme() = %q, want nord", selected.Name)
	}

	// 6. Test setting invalid theme returns error
	if err := SetTheme("invalid_theme_name"); err == nil {
		t.Errorf("expected SetTheme(invalid_theme_name) to error, got nil")
	}

	// 7. Test corrupt JSON fallback
	if err := os.WriteFile(filePath, []byte("{ invalid json }"), 0600); err != nil {
		t.Fatalf("writing corrupt json failed: %v", err)
	}
	corruptCfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig on corrupt file returned error: %v", err)
	}
	if corruptCfg.Theme != DefaultThemeName {
		t.Errorf("LoadConfig on corrupt file returned Theme = %q, want %q", corruptCfg.Theme, DefaultThemeName)
	}
}

func TestGetSelectedThemeWithEnvOverride(t *testing.T) {
	tempDir := t.TempDir()
	restore := SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	// Set saved config to gruvbox
	_ = SetTheme("gruvbox")

	// Without env, returns gruvbox
	t.Setenv("GHEPPO_THEME", "")
	if got := GetSelectedThemeName(); got != "gruvbox" {
		t.Errorf("GetSelectedThemeName() = %q, want gruvbox", got)
	}

	// With GHEPPO_THEME=catppuccin, env takes precedence
	t.Setenv("GHEPPO_THEME", "catppuccin")
	if got := GetSelectedThemeName(); got != "catppuccin" {
		t.Errorf("GetSelectedThemeName() with GHEPPO_THEME=catppuccin = %q, want catppuccin", got)
	}

	// With invalid GHEPPO_THEME, falls back to config (gruvbox)
	t.Setenv("GHEPPO_THEME", "invalid_custom_theme")
	if got := GetSelectedThemeName(); got != "gruvbox" {
		t.Errorf("GetSelectedThemeName() with invalid GHEPPO_THEME = %q, want gruvbox", got)
	}
}

func containsSubstring(s, sub string) bool {
	return filepath.Clean(s) != "" && len(s) >= len(sub) && (s == sub || (len(s) > len(sub) && findSub(s, sub)))
}

func findSub(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
