package config

import (
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func parseHex(hex string) (r, g, b float64, ok bool) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, 0, 0, false
	}
	val, err := strconv.ParseUint(hex, 16, 32)
	if err != nil {
		return 0, 0, 0, false
	}
	toLinear := func(c uint8) float64 {
		v := float64(c) / 255.0
		if v <= 0.04045 {
			return v / 12.92
		}
		return math.Pow((v+0.055)/1.055, 2.4)
	}
	return toLinear(uint8(val >> 16)), toLinear(uint8((val >> 8) & 0xFF)), toLinear(uint8(val & 0xFF)), true
}

func relativeLuminance(hex string) (float64, bool) {
	r, g, b, ok := parseHex(hex)
	if !ok {
		return 0, false
	}
	return 0.2126*r + 0.7152*g + 0.0722*b, true
}

func TestThemeEmptyCellContrastAndProgression(t *testing.T) {
	themes := ListThemes()
	seenEmptyColors := make(map[string]string)

	for _, theme := range themes {
		t.Run(theme.Name, func(t *testing.T) {
			// 1. Level-0 must not match the theme background (would make empty cells invisible)
			if strings.EqualFold(theme.GraphLevels[0], theme.Background) {
				t.Errorf("theme %q Level-0 color (%s) is identical to Background (%s); empty cells will be invisible",
					theme.Name, theme.GraphLevels[0], theme.Background)
			}
			if strings.EqualFold(theme.GraphEmpty, theme.Background) {
				t.Errorf("theme %q GraphEmpty color (%s) is identical to Background (%s)",
					theme.Name, theme.GraphEmpty, theme.Background)
			}

			// 2. GraphEmpty and GraphLevels[0] must match
			if !strings.EqualFold(theme.GraphEmpty, theme.GraphLevels[0]) {
				t.Errorf("theme %q GraphEmpty (%s) != GraphLevels[0] (%s)",
					theme.Name, theme.GraphEmpty, theme.GraphLevels[0])
			}

			// 3. Level-0 must be distinct across themes (theme-aware, not hardcoded global)
			if prevTheme, exists := seenEmptyColors[strings.ToLower(theme.GraphLevels[0])]; exists {
				t.Errorf("theme %q shares identical Level-0 color %s with theme %q; level-0 should be theme-aware",
					theme.Name, theme.GraphLevels[0], prevTheme)
			}
			seenEmptyColors[strings.ToLower(theme.GraphLevels[0])] = theme.Name

			// 4. All 5 graph levels must be pairwise distinct
			for i := 0; i < 5; i++ {
				for j := i + 1; j < 5; j++ {
					if strings.EqualFold(theme.GraphLevels[i], theme.GraphLevels[j]) {
						t.Errorf("theme %q has duplicate graph level colors at level %d and %d: %s",
							theme.Name, i, j, theme.GraphLevels[i])
					}
				}
			}

			// 5. Monotonic perceptual progression: Level 0 must be darker than Level 1, and so on
			var prevLum float64 = -1
			for lvl := 0; lvl < 5; lvl++ {
				lum, ok := relativeLuminance(theme.GraphLevels[lvl])
				if !ok {
					t.Fatalf("theme %q has invalid hex color at level %d: %s", theme.Name, lvl, theme.GraphLevels[lvl])
				}
				if lvl > 0 && lum <= prevLum {
					t.Errorf("theme %q level %d (lum=%.4f, %s) is not brighter than level %d (lum=%.4f, %s)",
						theme.Name, lvl, lum, theme.GraphLevels[lvl], lvl-1, prevLum, theme.GraphLevels[lvl-1])
				}
				prevLum = lum
			}
		})
	}
}

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

func TestSourceConfig(t *testing.T) {
	tempDir := t.TempDir()
	restore := SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	// Default source without config
	if src := GetSource(); src != SourceGitHub {
		t.Errorf("GetSource() default = %q, want %q", src, SourceGitHub)
	}
	if IsSourceConfigured() {
		t.Error("IsSourceConfigured() should be false initially")
	}

	// Persist source = leetcode
	if err := SetSource(SourceLeetCode); err != nil {
		t.Fatalf("SetSource(leetcode) failed: %v", err)
	}
	if !IsSourceConfigured() {
		t.Error("IsSourceConfigured() should be true after SetSource")
	}
	if src := GetSource(); src != SourceLeetCode {
		t.Errorf("GetSource() = %q, want %q", src, SourceLeetCode)
	}

	// LeetCode username persistence
	if err := SetLeetCodeUsername("test_user"); err != nil {
		t.Fatalf("SetLeetCodeUsername failed: %v", err)
	}
	if u := GetLeetCodeUsername(); u != "test_user" {
		t.Errorf("GetLeetCodeUsername() = %q, want test_user", u)
	}

	// Invalid source returns error
	if err := SetSource("invalid_source"); err == nil {
		t.Error("expected error for invalid source, got nil")
	}

	// Environment variable overrides
	t.Setenv("GHEPPO_SOURCE", "github")
	if src := GetSource(); src != SourceGitHub {
		t.Errorf("GetSource() with GHEPPO_SOURCE=github = %q, want %q", src, SourceGitHub)
	}

	t.Setenv("GHEPPO_LEETCODE_USERNAME", "env_user")
	if u := GetLeetCodeUsername(); u != "env_user" {
		t.Errorf("GetLeetCodeUsername() with env = %q, want env_user", u)
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
