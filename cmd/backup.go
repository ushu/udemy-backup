package cmd

import (
	"context"
	"os"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ushu/udemy-backup/backup"
	"github.com/ushu/udemy-backup/backup/config"
	"github.com/ushu/udemy-backup/backup/pool"
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

	// we can now load the generic backup options
	c := client.New(id, token)
	cfg := config.New(ctx, c)
	ctx = config.NewContext(ctx, cfg)

	// and prepare the worker pool
	workerPool := pool.New(cfg.NumWorkers)
	workerPool.RetryCount = 2 // retry 2 times on download failure
	ctx = pool.NewContext(ctx, workerPool)

	// here we start "enqueuing" the work on the pool
	go func() {
		defer workerPool.Done()

		if All {
			backupAllCourses(ctx, c)
		} else {
			if len(args) > 1 {
				courseID, err := strconv.Atoi(args[0])
				if err != nil {
					cli.Logerr("COURSE_ID should be a number (integer)")
					os.Exit(1)
				}
				backupCourse(ctx, c, courseID)
			} else {
				selectAndBackupCourse(ctx, c)
			}
		}
	}()

	if err := workerPool.Start(ctx); err != nil {
		cli.Logerr("Backup failed:", err)
		os.Exit(1)
	}
}

func backupAllCourses(ctx context.Context, c *client.Client) {
	// list all the course
	courses, err := c.ListAllCourses()
	if err != nil {
		cli.Logerrf("Failed to list courses: %v\n", err)
		os.Exit(1)
	}
	cli.Logf("⚙️  Found %d courses to backup\n", len(courses))

	for _, course := range courses {
		cli.Log("⚙️  Starting backup for:", course.Title)
		if err = backup.BackupCourse(ctx, course); err != nil {
			os.Exit(1)
		}
	}
}

func selectAndBackupCourse(ctx context.Context, c *client.Client) {
	// list all the course
	courses, err := c.ListAllCourses()
	if err != nil {
		cli.Logerrf("Failed to list courses: %v\n", err)
		os.Exit(1)
	}

	// prompt the user to select a course
	course, err := cli.SelectCourse(courses)
	if err != nil {
		cli.Logerrf("Could not select course: %v\n", err)
		os.Exit(1)
	}

	if err = backup.BackupCourse(ctx, course); err != nil {
		cli.Logerr("Backup failed:", err)
		os.Exit(1)
	}
}

func backupCourse(ctx context.Context, c *client.Client, courseID int) {
	course, err := c.GetCourse(courseID)
	if err != nil {
		cli.Logerr("Could not load course info:", err)
		os.Exit(1)
	}

	if err = backup.BackupCourse(ctx, course); err != nil {
		cli.Logerr("Backup failed:", err)
		os.Exit(1)
	}
}
