package cmd

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dexisback/gheppo/internal/config"
)

func TestThemeCommandList(t *testing.T) {
	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"theme"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("theme command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Current theme: github") {
		t.Errorf("output missing current theme: %q", out)
	}
	for _, expected := range []string{"github", "mono", "catppuccin", "nord", "gruvbox"} {
		if !strings.Contains(out, expected) {
			t.Errorf("output missing available theme %q: %q", expected, out)
		}
	}
}

func TestThemeCommandSetValid(t *testing.T) {
	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"theme", "nord"})

	if err := rootCmd.Execute(); err != nil {
		t.Fatalf("theme command failed: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "Theme set to 'nord'.") {
		t.Errorf("unexpected output: %q", out)
	}

	if current := config.GetSelectedThemeName(); current != "nord" {
		t.Errorf("config.GetSelectedThemeName() = %q, want nord", current)
	}
}

func TestThemeCommandSetInvalid(t *testing.T) {
	tempDir := t.TempDir()
	restore := config.SetConfigDirForTesting(func() (string, error) {
		return tempDir, nil
	})
	defer restore()

	buf := new(bytes.Buffer)
	rootCmd.SetOut(buf)
	rootCmd.SetErr(buf)
	rootCmd.SetArgs([]string{"theme", "nonexistent_theme"})

	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error setting nonexistent theme, got nil")
	}

	if !strings.Contains(err.Error(), "unknown theme") {
		t.Errorf("unexpected error message: %v", err)
	}
}

