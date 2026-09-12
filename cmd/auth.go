package cmd

import (
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


var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove GitHub authentication",
	RunE:  runLogout,
}


func init(){
	rootCmd.AddCommand(authCmd)
	rootCmd.AddCommand(loginCmd)
	rootCmd.AddCommand(logoutCmd)
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


