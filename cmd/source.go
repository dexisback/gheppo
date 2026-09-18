package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/config"
	"github.com/dexisback/gheppo/internal/leetcode"
	"github.com/dexisback/gheppo/internal/render"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var sourceCmd = &cobra.Command{
	Use:   "source [name] [username]",
	Short: "View or set the contribution data source",
	Long:  "Display the interactive source selector, or switch to a specific source (github, leetcode).",
	Args:  cobra.MaximumNArgs(2),
	RunE:  runSource,
}

func init() {
	rootCmd.AddCommand(sourceCmd)
}

func runSource(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	// Direct CLI arguments provided
	if len(args) > 0 {
		targetSource := strings.ToLower(strings.TrimSpace(args[0]))
		switch targetSource {
		case config.SourceGitHub:
			if err := config.SetSource(config.SourceGitHub); err != nil {
				return err
			}
			fmt.Fprintf(out, "Source set to 'github'.\n")
			if summary, ok := cache.LoadForSource(config.SourceGitHub); ok {
				render.Animate(out, summary)
			}
			return nil

		case config.SourceLeetCode:
			var username string
			if len(args) > 1 {
				username = strings.TrimSpace(args[1])
			} else {
				username = config.GetLeetCodeUsername()
			}
			return setupLeetCode(username, in, out)

		default:
			return fmt.Errorf("unknown source %q (available: %s)", targetSource, strings.Join(config.AvailableSources(), ", "))
		}
	}

	// Interactive mode in terminal
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

	// Non-interactive fallback: list sources
	current := config.GetSource()
	sources := config.AvailableSources()

	fmt.Fprintf(out, "Current source: %s\n\n", current)
	fmt.Fprintln(out, "Available sources:")
	for _, s := range sources {
		if s == current {
			fmt.Fprintf(out, "  • %s (active)\n", s)
		} else {
			fmt.Fprintf(out, "  • %s\n", s)
		}
	}
	fmt.Fprintln(out, "\nTo switch source, run:")
	fmt.Fprintln(out, "  gheppo source <name>")
	return nil
}

func handleSourceSelection(source string, in io.Reader, out io.Writer) error {
	switch source {
	case config.SourceGitHub:
		if err := config.SetSource(config.SourceGitHub); err != nil {
			return err
		}

		if summary, ok := cache.LoadForSource(config.SourceGitHub); ok {
			render.Animate(out, summary)
			return nil
		}

		fmt.Fprintln(out, "Source set to 'github'.")
		fmt.Fprintln(out)
		fmt.Fprintln(out, "Run:")
		fmt.Fprintln(out, "  gheppo auth login")
		fmt.Fprintln(out, "  gheppo sync")
		return nil

	case config.SourceLeetCode:
		existingUser := config.GetLeetCodeUsername()
		if existingUser != "" {
			if err := config.SetSource(config.SourceLeetCode); err != nil {
				return err
			}
			if summary, ok := cache.LoadForSource(config.SourceLeetCode); ok {
				render.Animate(out, summary)
				return nil
			}
		}

		return promptAndSetupLeetCode(in, out)

	default:
		return fmt.Errorf("unsupported source: %s", source)
	}
}

func setupLeetCode(username string, in io.Reader, out io.Writer) error {
	if username == "" {
		return promptAndSetupLeetCode(in, out)
	}

	fmt.Fprintf(out, "Validating @%s on LeetCode...\n", username)

	client := leetcode.NewClient()
	summary, err := client.FetchUserSummary(username)
	if err != nil {
		return fmt.Errorf("fetching leetcode user: %w", err)
	}

	if err := config.SetSource(config.SourceLeetCode); err != nil {
		return err
	}
	if err := config.SetLeetCodeUsername(username); err != nil {
		return err
	}
	if err := cache.SaveForSource(config.SourceLeetCode, summary); err != nil {
		return fmt.Errorf("saving cache: %w", err)
	}

	fmt.Fprintf(out, "Source set to 'leetcode' (@%s).\n\n", username)
	render.Animate(out, summary)
	return nil
}

func promptAndSetupLeetCode(in io.Reader, out io.Writer) error {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}

	fmt.Fprintln(out)
	fmt.Fprint(out, "Enter your LeetCode username: ")

	scanner := bufio.NewScanner(in)
	if !scanner.Scan() {
		return fmt.Errorf("missing LeetCode username: run `gheppo source leetcode <username>`")
	}

	username := strings.TrimSpace(scanner.Text())
	if username == "" {
		return fmt.Errorf("username cannot be empty")
	}

	return setupLeetCode(username, in, out)
}
