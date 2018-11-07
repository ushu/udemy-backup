package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/ushu/udemy-backup/cli"
	"github.com/ushu/udemy-backup/client"
)

// loginCmd represents the login command
var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Tries to login to the Udemy account",
	Long:  `Attempts to log in to Udemy using the provided credentials, and reports the status.`,
	Run:   login,
}

func init() {
	rootCmd.AddCommand(loginCmd)
}

func login(cmd *cobra.Command, args []string) {
	// grab credentials
	id, token, err := cli.EnsureCredentials()
	if err != nil {
		cli.Logerrf("Failed to load credentials: %v\n", err)
		os.Exit(1)
	}

	// and test connection to the remote server
	c := client.New(id, token)
	user, err := c.GetUser()
	if err != nil {
		cli.Logerrf("Failed to user info: %v\n", err)
		os.Exit(1)
	}

	cli.Log()
	cli.Log("🍾  SUCCESSFULLY AUTHENTICATED WITH UDEMY")
	cli.Log("🍾  User name:", user.DisplayName)
	cli.Log("🍾  Udemy ID :", user.ID)
}
