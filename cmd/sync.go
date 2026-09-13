package cmd

import (
	"fmt"
	"os"

	"github.com/dexisback/gheppo/internal/auth"
	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/github"
	"github.com/dexisback/gheppo/internal/stats"
	"github.com/spf13/cobra"
)

const refreshLockTokenEnv = "GHEPPO_REFRESH_LOCK_TOKEN"

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Fetch your latest GitHub contribution data",
	RunE:  runSync,
}

// runs even before main()
func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// A manually executed `gheppo sync` does not have this
	// environment variable, while a detached background refresh does.
	background := os.Getenv("GHEPPO_BACKGROUND_REFRESH") == "1"

	if background {
		lockToken := os.Getenv(refreshLockTokenEnv)

		// Always attempt to release the lock when the background
		// refresh finishes, whether the sync succeeds or fails.
		defer cache.ReleaseRefreshLockWithToken(lockToken)
	}

	credentials, err := auth.GetCredentials()
	if err != nil {
		return fmt.Errorf(
			"not logged in — run `gheppo auth login` first: %w",
			err,
		)
	}

	client := github.NewClient(credentials.Token)

	cal, err := client.FetchContributionCalendar(credentials.Login)
	if err != nil {
		return fmt.Errorf("fetching contributions: %w", err)
	}

	summary := stats.Summarize(cal)

	if err := cache.Save(summary); err != nil {
		return fmt.Errorf("saving cache: %w", err)
	}

	fmt.Fprintf(
		cmd.OutOrStdout(),
		"synced %d contributions for @%s\n",
		summary.Total,
		summary.Login,
	)

	return nil
}