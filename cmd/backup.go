package cmd

import (
	"context"
	"os"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ushu/udemy-backup/backup"
	"github.com/ushu/udemy-backup/cli"
	"github.com/ushu/udemy-backup/client"
)

var PreferredResolution int
var NumWorkers int
var Dir string
var Restart bool
var All bool
var Subtitles bool

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup [COURSE_ID]",
	Short: "Backup a course",
	Long:  `Downloads a backup for a course, given by its URL or selected amoung all the subscribed courses.`,
	Run:   runBackup,
}

func init() {
	rootCmd.AddCommand(backupCmd)
	backupCmd.PersistentFlags().IntVar(&PreferredResolution, "resolution", 0, "only download videos of the given resolution")
	backupCmd.PersistentFlags().IntVar(&NumWorkers, "concurrency", runtime.NumCPU(), "number of parallel downloads")
	backupCmd.PersistentFlags().StringVar(&Dir, "dir", ".", "output directory for downloads")
	backupCmd.PersistentFlags().BoolVar(&Restart, "restart", false, "skip download of existing files")
	backupCmd.PersistentFlags().BoolVar(&All, "all", false, "backup all the subscribed courses for the account")
	backupCmd.PersistentFlags().BoolVar(&Subtitles, "subtitles", false, "download subtitles (vtt) files")
	viper.BindPFlag("resolution", backupCmd.PersistentFlags().Lookup("resolution"))
	viper.BindPFlag("concurrency", backupCmd.PersistentFlags().Lookup("concurrency"))
	viper.BindPFlag("dir", backupCmd.PersistentFlags().Lookup("dir"))
	viper.BindPFlag("restart", backupCmd.PersistentFlags().Lookup("restart"))
	viper.BindPFlag("subtitles", backupCmd.PersistentFlags().Lookup("subtitles"))
}

func runBackup(cmd *cobra.Command, args []string) {
	ctx := context.Background()

	// grab credentials
	id, token, err := cli.EnsureCredentials()
	if err != nil {
		cli.Logerr("Failed to load credentials: %v\n", err)
		os.Exit(1)
	}

	// we can now connect to Udemy
	c := client.New(id, token)
	ctx = backup.SetClient(ctx, c)

	if All {
		// list all the course
		courses, err := c.ListAllCourses()
		if err != nil {
			cli.Logerrf("Failed to list courses: %v\n", err)
			os.Exit(1)
		}
		cli.Logf("⚙️  Found %d courses to backup\n", len(courses))

		for _, course := range courses {
			cli.Log("⚙️  Starting backup for:", course.Title)
			err = backupCourse(ctx, course)
			if err != nil {
				os.Exit(1)
			}
		}
	} else {
		var course *client.Course
		if len(args) > 0 {
			courseID, err := strconv.Atoi(args[0])
			if err != nil {
				cli.Logerr("COURSE_ID should be a number (integer)")
			}
			course, err = c.GetCourse(courseID)
			if err != nil {
				cli.Logerr("Could not load course info:", err)
				os.Exit(1)
			}
		} else {
			// list all the course
			courses, err := c.ListAllCourses()
			if err != nil {
				cli.Logerrf("Failed to list courses: %v\n", err)
				os.Exit(1)
			}

			// prompt the user to select a course
			course, err = cli.SelectCourse(courses)
			if err != nil {
				cli.Logerrf("Could not select course: %v\n", err)
				os.Exit(1)
			}
		}

		// backup starts here
		err = backupCourse(ctx, course)
		if err != nil {
			os.Exit(1)
		}
	}
}

func backupCourse(ctx context.Context, course *client.Course) error {
	err := backup.Run(ctx, course)
	if err == nil {
		cli.Log("🍾 Done backuping course", course.Title)
	} else {
		cli.Logerr("☠️  Error while backuping course:", err)
	}
	return err
}
