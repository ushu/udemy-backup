package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/ushu/udemy-backup/cli"
)

var cfgFile string
var ID string
var Quiet bool
var Token string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "udemy-backup",
	Short: "Create backups of subscribed Udemy courses",
	Long:  `A tool that create backups of Udemy courses, given API crendentials and a course URL.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		cli.Logerr(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.udemy-backup)")

	// credentials
	rootCmd.PersistentFlags().StringVar(&ID, "id", "", "Udemy ID for the user")
	rootCmd.PersistentFlags().StringVar(&Token, "token", "", "Udemy Access Token for the user")
	rootCmd.PersistentFlags().BoolVar(&Quiet, "quiet", false, "Reduce additional into")
	viper.BindPFlag("id", rootCmd.PersistentFlags().Lookup("id"))
	viper.BindPFlag("token", rootCmd.PersistentFlags().Lookup("token"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath("$HOME")
		viper.SetConfigName(".udemy-backup")
	}

	viper.SetEnvPrefix("UDEMY")
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		cli.Log("Using config file:", viper.ConfigFileUsed())
	}
}
