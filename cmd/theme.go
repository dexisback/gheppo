package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/config"
	"github.com/dexisback/gheppo/internal/render"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var themeCmd = &cobra.Command{
	Use:   "theme [name]",
	Short: "View or set the visual theme",
	Long:  "Display the interactive theme selector, or switch to a specific theme.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTheme,
}

func init() {
	rootCmd.AddCommand(themeCmd)
}

func runTheme(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// If a theme name was directly provided via arguments, apply it directly
	if len(args) > 0 {
		target := strings.ToLower(strings.TrimSpace(args[0]))
		if err := config.SetTheme(target); err != nil {
			return err
		}

		fmt.Fprintf(out, "Theme set to '%s'.\n", target)
		return nil
	}

	// Interactive mode: if running in a terminal, open the Bubble Tea interactive selector
	if term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) {
		chosen, canceled, err := render.RunThemeSelector(cmd.InOrStdin(), out)
		if err != nil {
			return err
		}
		if canceled {
			return nil
		}

		if err := config.SetTheme(chosen.Name); err != nil {
			return err
		}

		if summary, ok := cache.Load(); ok {
			render.Animate(out, summary)
		} else {
			fmt.Fprintf(out, "Theme set to '%s'.\n", chosen.Name)
		}
		return nil
	}

	// Non-interactive fallback: display current theme and available themes
	current := config.GetSelectedThemeName()
	themes := config.ListThemes()

	fmt.Fprintf(out, "Current theme: %s\n\n", current)
	fmt.Fprintln(out, "Available themes:")
	for _, t := range themes {
		if t.Name == current {
			fmt.Fprintf(out, "  • %s (active)\n", t.Name)
		} else {
			fmt.Fprintf(out, "  • %s\n", t.Name)
		}
	}
	fmt.Fprintln(out, "\nTo switch theme, run:")
	fmt.Fprintln(out, "  gheppo theme <name>")
	return nil
}
