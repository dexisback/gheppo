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
	//add lock release (because the detached sync process eventually will say "im done, release the lock")
	background := os.Getenv("GHEPPO_BACKGROUND_REFRESH") == "1"
	if background {
		defer cache.ReleaseRefreshLock()
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
