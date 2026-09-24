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
[[ -o interactive ]] || return 0

if [[ -o zle ]]; then
    autoload -Uz add-zle-hook-widget

    _gheppo_once() {
        add-zle-hook-widget -d line-init _gheppo_once
        zle && zle -I
        command gheppo
    }

    add-zle-hook-widget line-init _gheppo_once
else
    command gheppo
fi
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

func TestRemoveShellIntegrationRemovesPathAndIntegrationBlocks(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")

	content := `alias ll='ls -lah'

# >>> gheppo PATH >>>
export PATH="/home/u/.local/bin:$PATH"
# <<< gheppo PATH <<<

# >>> gheppo >>>
command gheppo
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

	for _, marker := range []string{gheppoPathStart, gheppoPathEnd, gheppoShellStart, gheppoShellEnd} {
		if strings.Contains(result, marker) {
			t.Fatalf("marker %q still exists", marker)
		}
	}

	if !strings.Contains(result, `alias ll='ls -lah'`) {
		t.Fatal("user configuration before the Gheppo blocks was removed")
	}

	if !strings.Contains(result, `export EDITOR=vim`) {
		t.Fatal("user configuration after the Gheppo blocks was removed")
	}
}

func TestRemoveShellIntegrationRemovesPathOnlyBlock(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")

	content := `export PATH="$HOME/.local/bin:$PATH"

# >>> gheppo PATH >>>
export PATH="/home/u/.local/bin:$PATH"
# <<< gheppo PATH <<<

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

	if strings.Contains(result, gheppoPathStart) || strings.Contains(result, gheppoPathEnd) {
		t.Fatal("PATH block was not removed")
	}

	if !strings.Contains(result, `export PATH="$HOME/.local/bin:$PATH"`) {
		t.Fatal("user's own PATH export was removed")
	}
}

func TestRemoveShellIntegrationLeavesIncompletePathBlock(t *testing.T) {
	dir := t.TempDir()
	rc := filepath.Join(dir, ".zshrc")

	content := `# >>> gheppo PATH >>>
export PATH="/home/u/.local/bin:$PATH"
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
		t.Fatal("incomplete PATH block should not have been removed")
	}
}

func TestRemovePathEntry(t *testing.T) {
	cases := []struct {
		name      string
		pathValue string
		dir       string
		want      string
		changed   bool
	}{
		{
			name:      "removes matching entry",
			pathValue: `C:\a;C:\Users\u\.local\bin;C:\b`,
			dir:       `C:\Users\u\.local\bin`,
			want:      `C:\a;C:\b`,
			changed:   true,
		},
		{
			name:      "case-insensitive match",
			pathValue: `c:\users\u\.local\bin`,
			dir:       `C:\Users\u\.local\bin`,
			want:      "",
			changed:   true,
		},
		{
			name:      "trims surrounding whitespace",
			pathValue: `C:\a; C:\Users\u\.local\bin ;C:\b`,
			dir:       `C:\Users\u\.local\bin`,
			want:      `C:\a;C:\b`,
			changed:   true,
		},
		{
			name:      "dir not present",
			pathValue: `C:\a;C:\b`,
			dir:       `C:\x`,
			want:      `C:\a;C:\b`,
			changed:   false,
		},
		{
			name:      "only entry becomes empty",
			pathValue: `C:\Users\u\.local\bin`,
			dir:       `C:\Users\u\.local\bin`,
			want:      "",
			changed:   true,
		},
		{
			name:      "partial match is kept",
			pathValue: `C:\Users\u\.local\bin2`,
			dir:       `C:\Users\u\.local\bin`,
			want:      `C:\Users\u\.local\bin2`,
			changed:   false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, changed := removePathEntry(tc.pathValue, tc.dir)
			if got != tc.want {
				t.Fatalf("removePathEntry() = %q, want %q", got, tc.want)
			}
			if changed != tc.changed {
				t.Fatalf("changed = %v, want %v", changed, tc.changed)
			}
		})
	}
}
