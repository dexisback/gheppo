package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/dexisback/gheppo/internal/auth"
	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/config"
	"github.com/dexisback/gheppo/internal/github"
	"github.com/dexisback/gheppo/internal/render"
	"github.com/dexisback/gheppo/internal/stats"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage authentication and data sources",
}

var loginCmd = &cobra.Command{
	Use:   "login [source] [username]",
	Short: "Authenticate with GitHub or configure LeetCode source",
	Args:  cobra.MaximumNArgs(2),
	RunE:  runLogin,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show authentication and data source status",
	RunE:  runStatus,
}

var logoutCmd = &cobra.Command{
	Use:   "logout [source]",
	Short: "Remove authentication or clear source configuration",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)
}

func runLogin(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	in := cmd.InOrStdin()

	if len(args) > 0 {
		targetSource := strings.ToLower(strings.TrimSpace(args[0]))
		switch targetSource {
		case config.SourceGitHub:
			return setupGitHub(in, out)

		case config.SourceLeetCode:
			var username string
			if len(args) > 1 {
				username = strings.TrimSpace(args[1])
			}
			return setupLeetCode(username, in, out)

		default:
			return fmt.Errorf("unknown source %q (available: %s)", targetSource, strings.Join(config.AvailableSources(), ", "))
		}
	}

	// Interactive mode: launch source selector
	if term.IsTerminal(int(os.Stdout.Fd())) && term.IsTerminal(int(os.Stdin.Fd())) {
		chosen, canceled, err := render.RunSourceSelector(in, out)
		if err != nil {
			return err
		}
		if canceled || chosen == "" {
			return nil
		}

		switch chosen {
		case config.SourceGitHub:
			return setupGitHub(in, out)
		case config.SourceLeetCode:
			return promptAndSetupLeetCode(in, out)
		default:
			return fmt.Errorf("unsupported source: %s", chosen)
		}
	}

	// Non-interactive fallback: if source is already LeetCode, prompt LeetCode, otherwise GitHub
	if config.GetSource() == config.SourceLeetCode {
		return promptAndSetupLeetCode(in, out)
	}
	return setupGitHub(in, out)
}

func setupGitHub(in io.Reader, out io.Writer) error {
	creds, err := auth.PromptAndLogin(in, out)
	if err != nil {
		return fmt.Errorf("login failed: %w", err)
	}

	if err := config.SetSource(config.SourceGitHub); err != nil {
		return err
	}

	fmt.Fprintln(out, "Fetching your GitHub contribution data...")

	client := github.NewClient(creds.Token)
	cal, err := client.FetchContributionCalendar(creds.Login)
	if err != nil {
		return fmt.Errorf("fetching contributions: %w", err)
	}

	summary := stats.Summarize(cal)
	if err := cache.SaveForSource(config.SourceGitHub, summary); err != nil {
		return fmt.Errorf("saving cache: %w", err)
	}

	fmt.Fprintln(out)
	render.Animate(out, summary)
	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	source := config.GetSource()
	if len(args) > 0 {
		source = strings.ToLower(strings.TrimSpace(args[0]))
	}

	switch source {
	case config.SourceLeetCode:
		if err := config.SetLeetCodeUsername(""); err != nil {
			return fmt.Errorf("clearing leetcode username: %w", err)
		}
		fmt.Fprintln(out, "Logged out of LeetCode.")
		return nil

	case config.SourceGitHub:
		if err := auth.ClearToken(); err != nil {
			return fmt.Errorf("logout failed: %w", err)
		}
		fmt.Fprintln(out, "Logged out of GitHub.")
		return nil

	default:
		if err := auth.ClearToken(); err != nil {
			return fmt.Errorf("logout failed: %w", err)
		}
		fmt.Fprintln(out, "Logged out of GitHub.")
		return nil
	}
}

func runStatus(cmd *cobra.Command, args []string) error {
	out := cmd.OutOrStdout()
	source := config.GetSource()

	if source == config.SourceLeetCode {
		username := config.GetLeetCodeUsername()
		if username == "" {
			fmt.Fprintln(out, "Not logged in to leetcode (no username configured)")
			return nil
		}
		fmt.Fprintf(out, "Active source: leetcode (@%s)\n", username)
		return nil
	}

	credentials, err := auth.GetCredentials()
	if err != nil {
		if errors.Is(err, auth.ErrNoToken) {
			fmt.Fprintln(out, "Not logged in to github")
			return nil
		}
		return fmt.Errorf("checking authentication status: %w", err)
	}
	fmt.Fprintf(out, "Logged in to github as @%s.\n", credentials.Login)
	return nil
}
