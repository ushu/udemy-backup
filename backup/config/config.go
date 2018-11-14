package config

import (
	"context"

	"github.com/spf13/viper"
	"github.com/ushu/udemy-backup/client"
)

type Config struct {
	Client        *client.Client
	Resolution    int
	NumWorkers    int
	Restart       bool
	LoadSubtitles bool
	RootDir       string
}

func New(ctx context.Context, c *client.Client) *Config {
	return &Config{
		Client:        c,
		Resolution:    viper.GetInt("resolution"),
		NumWorkers:    viper.GetInt("concurrency"),
		Restart:       viper.GetBool("restart"),
		LoadSubtitles: viper.GetBool("subtitles"),
		RootDir:       viper.GetString("dir"),
	}
}
