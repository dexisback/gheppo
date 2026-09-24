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
	gheppoPathStart  = "# >>> gheppo PATH >>>"
	gheppoPathEnd    = "# <<< gheppo PATH <<<"
)

// managedBlock is a marker-delimited region written to a shell rc file.
type managedBlock struct {
	start string
	end   string
}

// managedBlocks lists every block scripts/install.sh appends to shell rc
// files, so uninstall reverses all of them.
var managedBlocks = []managedBlock{
	{gheppoPathStart, gheppoPathEnd},
	{gheppoShellStart, gheppoShellEnd},
}

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove Gheppo from your system",
	RunE:  runUninstall,
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}

func runUninstall(cmd *cobra.Command, args []string) error {
	// Shell integration cleanup is best-effort so unsupported shells (e.g. a
	// native Windows install without $SHELL) can still complete the rest of
	// the uninstall flow.
	if shellRC, err := shellRCPath(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Skipping shell integration cleanup: %v\n", err)
	} else if err := removeShellIntegration(shellRC); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Skipping shell integration cleanup: %v\n", err)
	}

	// Reverses the Windows user PATH entry added by scripts/install.sh.
	if err := cleanupUserPath(); err != nil {
		fmt.Fprintf(cmd.ErrOrStderr(), "Skipping PATH cleanup: %v\n", err)
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

// removeShellIntegration removes every managed Gheppo block (persistent PATH
// entry and shell integration) from the given rc file.
func removeShellIntegration(path string) error {
	for _, block := range managedBlocks {
		if err := removeManagedBlock(path, block.start, block.end); err != nil {
			return err
		}
	}
	return nil
}

func removeManagedBlock(path, startMarker, endMarker string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	content := string(data)

	start := strings.Index(content, startMarker)
	if start == -1 {
		return nil
	}

	endOffset := strings.Index(content[start:], endMarker)
	if endOffset == -1 {
		return nil
	}

	end := start + endOffset + len(endMarker)

	content = content[:start] + content[end:]

	return os.WriteFile(path, []byte(content), 0644)
}
