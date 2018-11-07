package cmd

import (
	"encoding/csv"
	"os"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/ushu/udemy-backup/cli"
	"github.com/ushu/udemy-backup/client"
)

var CSV bool

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all the subscribed courses",
	Run:   list,
}

func init() {
	listCmd.PersistentFlags().BoolVar(&CSV, "csv", false, "Write output in CSV format")
	rootCmd.AddCommand(listCmd)
}

func list(cmd *cobra.Command, args []string) {
	// grab credentials
	id, token, err := cli.EnsureCredentials()
	if err != nil {
		cli.Logerrf("Failed to load credentials: %v\n", err)
		os.Exit(1)
	}

	// and test connection to the remote server
	c := client.New(id, token)
	courses, err := c.ListAllCourses()
	if err != nil {
		cli.Logerrf("Failed to list courses: %v\n", err)
		os.Exit(1)
	}

	if CSV {
		// output CSV format with ";" separator (for Exel)
		w := csv.NewWriter(os.Stdout)
		w.Comma = ';'
		w.Write([]string{"ID", "Title"})
		for _, course := range courses {
			if err := w.Write([]string{strconv.Itoa(course.ID), course.Title}); err != nil {
				break
			}
		}
		w.Flush()
	} else {
		cli.Logf("| %-7s | %-60s |\n", "ID", "Title")
		for _, course := range courses {
			cli.Logf("| %-7v | %-60s |\n", course.ID, course.Title)
		}
	}
}
