// this file determined what happens when someone runs the root command (gheppo, just gheppo and nothing else )
// cobra finally enters the picture
// cmd/sync.go and cmd/auth.go will each define their own cobra.Command and call rootCmd.AddCommand(...)
// init functions -- go runs any function named exactly init() automatically before main(), per package at program startup w/o the need of any explicit call.
// this is the mechanism which cobra uses to self-register. each subcommand file's init() does rootCmd.AddCommand(syncCmd)
// cobra commands are just a tree: rootCmd has children(syncCmd, authCmd) and authCmd could itself have children(loginCmd
package cmd

import (
	"fmt"
	"os"

	"github.com/dexisback/gheppo/internal/auth"
	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/config"
	"github.com/dexisback/gheppo/internal/refresh"
	"github.com/dexisback/gheppo/internal/render"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var version = "0.2.0"

var rootCmd = &cobra.Command{
	Use:     "gheppo",
	Short:   "Your contribution graph, everytime you open a terminal session",
	Version: version,
	RunE:    runDefault,
}

func Execute() error {
	return rootCmd.Execute()
}

func runDefault(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	// First-run experience: if source is not yet configured and running interactively,
	// launch the compact source selector (styled in gruvbox)
	if !config.IsSourceConfigured() {
		if term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) {
			chosen, canceled, err := render.RunSourceSelector(in, out)
			if err != nil {
				return err
			}
			if canceled || chosen == "" {
				return nil
			}

			return handleSourceSelection(chosen, in, out)
		}

		// Non-interactive fallback for first run: default to GitHub
		_ = config.SetSource(config.SourceGitHub)
	}

	source := config.GetSource()
	summary, ok := cache.LoadForSource(source)
	if !ok {
		if source == config.SourceLeetCode {
			username := config.GetLeetCodeUsername()
			if username == "" {
				if term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) {
					return promptAndSetupLeetCode(in, out)
				}
				fmt.Fprintln(out, "No LeetCode username configured.")
				fmt.Fprintln(out)
				fmt.Fprintln(out, "Run:")
				fmt.Fprintln(out, "  gheppo source leetcode <username>")
				return nil
			}

			fmt.Fprintln(out, "No contribution data found.")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Run:")
			fmt.Fprintln(out, "  gheppo sync")
			return nil
		}

		// GitHub source
		_, err := auth.GetCredentials()
		if err != nil {
			fmt.Fprintln(out, "Gheppo isn't set up yet.")
			fmt.Fprintln(out)
			fmt.Fprintln(out, "Run:")
			fmt.Fprintln(out, "  gheppo auth login")
			fmt.Fprintln(out, "  gheppo sync")
			return nil
		}

		fmt.Fprintln(out, "No contribution data found.")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Run:")
		fmt.Fprintln(out, "  gheppo sync")
		return nil
	}

	render.Animate(out, summary)

	_ = refresh.MaybeRefresh()
	return nil
}

