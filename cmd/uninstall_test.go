package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRemoveShellIntregration(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")

	content := `export PATH="$HOME/.local/bin:$PATH"

# user's configuration
alias ll='ls -lah'

# >>> gheppo >>>
autoload -Uz add-zsh-hook

_gheppo_once() {
	add-zsh-hook -d precmd _gheppo_once
	command gheppo
}

add-zsh-hook precmd _gheppo_once
# <<< gheppo <<<

export EDITOR=vim
`
	if err := os.WriteFile(rc, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := removeShellIntegration(rc); err != nil {
		t.Fatalf("removeShellIntegration() error = %v", err)
	}

	data, err := os.ReadFile(rc)
	if err != nil {
		t.Fatal(err)
	}

	result := string(data)

	if strings.Contains(result, gheppoShellStart) {
		t.Fatal("Gheppo start marker still exists")
	}

	if strings.Contains(result, gheppoShellEnd) {
		t.Fatal("Gheppo end marker still exists")
	}

	if !strings.Contains(result, `alias ll='ls -lah'`) {
		t.Fatal("user configuration was removed")
	}

	if !strings.Contains(result, `export EDITOR=vim`) {
		t.Fatal("configuration after Gheppo block was removed")
	}
}

func TestRemoveShellIntegrationWhenNotInstalled(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")

	content := `export PATH="$HOME/.local/bin:$PATH"
alias ll='ls -lah'
`

	if err := os.WriteFile(rc, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := removeShellIntegration(rc); err != nil {
		t.Fatalf("removeShellIntegration() error = %v", err)
	}

	data, err := os.ReadFile(rc)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != content {
		t.Fatal("file was modified even though Gheppo was not installed")
	}
}

func TestRemoveShellIntegrationWhenFileDoesNotExist(t *testing.T) {
	path := filepath.Join(t.TempDir(), ".zshrc")

	if err := removeShellIntegration(path); err != nil {
		t.Fatalf("removeShellIntegration() error = %v", err)
	}
}

func TestRemoveShellIntegrationWhenBlockIsIncomplete(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")

	content := `export PATH="$HOME/.local/bin:$PATH"

# >>> gheppo >>>
autoload -Uz add-zsh-hook
`

	if err := os.WriteFile(rc, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	if err := removeShellIntegration(rc); err != nil {
		t.Fatalf("removeShellIntegration() error = %v", err)
	}

	data, err := os.ReadFile(rc)
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != content {
		t.Fatal("incomplete Gheppo block should not have been removed")
	}
}
