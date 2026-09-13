// this file determined what happens when someone runs the root command (gheppo, just gheppo and nothing else )
// cobra finally enters the picture
// cmd/sync.go and cmd/auth.go will each define their own cobra.Command and call rootCmd.AddCommand(...)
// init functions -- go runs any function named exactly init() automatically before main(), per package at program startup w/o the need of any explicit call.
// this is the mechanism which cobra uses to self-register. each subcommand file's init() does rootCmd.AddCommand(syncCmd)
// cobra commands are just a tree: rootCmd has children(syncCmd, authCmd) and authCmd could itself have children(loginCmd
package cmd

import (
	"fmt"

	"github.com/dexisback/gheppo/internal/cache"
	"github.com/dexisback/gheppo/internal/refresh"
	"github.com/dexisback/gheppo/internal/render"

	"github.com/spf13/cobra"
	"github.com/dexisback/gheppo/internal/auth"
)

var version = "0.1.0"

var rootCmd = &cobra.Command{
	Use:     "gheppo",
	Short:   "Your github contribution graph, everytime you open a terminal session",
	Version: version,
	RunE:    runDefault,
}

func Execute() error {
	return rootCmd.Execute()
}

// func runDefault(cmd *cobra.Command, args []string) error {
// 	summary, ok := cache.Load()
// 	if !ok {
// 		fmt.Fprintln(
// 			cmd.OutOrStdout(),
// 			"Gheppo isn't set up yet.",
// 		)
// 		fmt.Fprintln(
// 			cmd.OutOrStdout(),
// 			"Run `gheppo auth login` then `gheppo sync`.",
// 		)
// 		return nil
// 	}

// 	output := render.Grid(summary)
// 	fmt.Fprintln(cmd.OutOrStdout(), output)

// 	//refresh is deliberately non-critical for the shell startup path
// 	//the cached graph has alr been rendered, so a refresh failure should never make `gheppo` fail
// 	_ = refresh.MaybeRefresh()

// 	return nil
// }


func runDefault(cmd *cobra.Command, args []string) error {
	summary, ok := cache.Load()
	if !ok {
		_, err := auth.GetCredentials()
		if err != nil {
			fmt.Fprintln(cmd.OutOrStdout(), "Gheppo isn't set up yet.")
			fmt.Fprintln(cmd.OutOrStdout())
			fmt.Fprintln(cmd.OutOrStdout(), "Run:")
			fmt.Fprintln(cmd.OutOrStdout(), "  gheppo auth login")
			fmt.Fprintln(cmd.OutOrStdout(), "  gheppo sync")
			return nil
		}

		fmt.Fprintln(cmd.OutOrStdout(), "No contribution data found.")
		fmt.Fprintln(cmd.OutOrStdout())
		fmt.Fprintln(cmd.OutOrStdout(), "Run:")
		fmt.Fprintln(cmd.OutOrStdout(), "  gheppo sync")
		return nil
	}

	output := render.Grid(summary)
	fmt.Fprintln(cmd.OutOrStdout(), output)

	_ = refresh.MaybeRefresh()
	return nil
}

