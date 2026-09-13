package cmd

import (
	"errors"
	"fmt"

	"github.com/dexisback/gheppo/internal/auth"
	"github.com/spf13/cobra"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage GitHub authentication",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with GitHub",
	RunE:  runLogin,
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show github authentication status",
	RunE:  runStatus,
}

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove GitHub authentication",
	RunE:  runLogout,
}

func init() {
	rootCmd.AddCommand(authCmd)
	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(logoutCmd)
	authCmd.AddCommand(statusCmd)

}

func runLogin(cmd *cobra.Command, args []string) error {
	if err := auth.Login(); err != nil {
		return fmt.Errorf("login failed: %w:", err)

	}
	return nil
}

func runLogout(cmd *cobra.Command, args []string) error {
	if err := auth.ClearToken(); err != nil {
		return fmt.Errorf("logout failed: %w", err)
	}
	fmt.Fprintln(cmd.OutOrStdout(), "Logged out of GitHub.")

	return nil
}

func runStatus(cmd *cobra.Command, args []string) error {
	credentials, err := auth.GetCredentials()
	if err != nil {
		if errors.Is(err, auth.ErrNoToken) {
			fmt.Fprintln(cmd.OutOrStdout(), "Not logged in to github")
			return nil
		}
		return fmt.Errorf("checking authentication status: %w", err)

	}
	fmt.Fprintf(cmd.OutOrStdout(), "Logged in to github as @%s.\n", credentials.Login)
	return nil
}
