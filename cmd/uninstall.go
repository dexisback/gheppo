package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dexisback/gheppo/internal/auth"
	"github.com/spf13/cobra"
)

// now the markers will come in handy :)
const (
	gheppoShellStart = "# >>> gheppo >>>"
	gheppoShellEnd   = "# <<< gheppo <<<"
)

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove Gheppo from your system",
	RunE:  runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	shellRC, err := shellRCPath()
	if err != nil {
		return err
	}
	if err := removeShellIntegration(shellRC); err != nil {
		return fmt.Errorf("removing shell integration: %w", err)
	}

	if err := auth.ClearToken(); err != nil {
		return fmt.Errorf("removing authentication: %w", err)
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("locating Gheppo executable: %w", err)
	}

	if err := os.Remove(executable); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("removing Gheppo executable: %w", err)
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Gheppo has been uninstalled.")
	return nil
}

func shellRCPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("locating home directory: %w", err)
	}

	switch filepath.Base(os.Getenv("SHELL")) {
	case "zsh":
		return filepath.Join(home, ".zshrc"), nil
	case "bash":
		return filepath.Join(home, ".bashrc"), nil
	default:
		return "", fmt.Errorf("unsupported shell: %s", os.Getenv("SHELL"))
	}
}

func removeShellIntegration(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	content := string(data)

	start := strings.Index(content, gheppoShellStart)
	if start == -1 {
		return nil
	}

	endOffset := strings.Index(content[start:], gheppoShellEnd)
	if endOffset == -1 {
		return nil
	}

	end := start + endOffset + len(gheppoShellEnd)

	content = content[:start] + content[end:]

	return os.WriteFile(path, []byte(content), 0644)
}
