package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemoveExecutableRemovesUnlockedFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "gheppo")
	if err := os.WriteFile(path, []byte("binary"), 0755); err != nil {
		t.Fatal(err)
	}

	if err := removeExecutable(path); err != nil {
		t.Fatalf("removeExecutable() error = %v", err)
	}

	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("executable was not removed")
	}
}

func TestRemoveExecutableMissingFile(t *testing.T) {
	if err := removeExecutable(filepath.Join(t.TempDir(), "gheppo")); err != nil {
		t.Fatalf("removeExecutable() error = %v", err)
	}
}
