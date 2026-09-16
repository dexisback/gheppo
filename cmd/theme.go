package cmd

import (
	"fmt"
	"strings"

	"github.com/dexisback/gheppo/internal/config"
	"github.com/spf13/cobra"
)

var themeCmd = &cobra.Command{
	Use:   "theme [name]",
	Short: "View or set the visual theme",
	Long:  "Display the current theme and available themes, or switch to a specific theme.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runTheme,
}

func init() {
	rootCmd.AddCommand(themeCmd)
}

func runTheme(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()

	// If no args provided, display current theme and available themes
	if len(args) == 0 {
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

	target := strings.ToLower(strings.TrimSpace(args[0]))
	if err := config.SetTheme(target); err != nil {
		return err
	}

	fmt.Fprintf(out, "Theme set to '%s'.\n", target)
	return nil
}
