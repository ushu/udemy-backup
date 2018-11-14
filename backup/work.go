package backup

import (
	"context"
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/ushu/udemy-backup/cli"
)

type WorkType int

const (
	WorkTypeDownload WorkType = iota
	WorkTypeWriteFile
)

type Work struct {
	Type    WorkType
	Config  *Config
	Payload interface{}
}

type DownloadPayload struct {
	URL      string
	FilePath string
}

type WriteFilePayload struct {
	Contents []byte
	FilePath string
}

func (w Work) Run(ctx context.Context) error {
	var err error
	fileName := "?"

	switch w.Type {
	case WorkTypeDownload:
		d := w.Payload.(*DownloadPayload)
		fileName = filepath.Base(d.FilePath)
		if !w.Config.Restart || !FileExists(d.FilePath) {
			err = downloadURLToFile(w.Config.Client, d.URL, d.FilePath)
		}
	case WorkTypeWriteFile:
		d := w.Payload.(*WriteFilePayload)
		fileName = filepath.Base(d.FilePath)
		cli.Logf("⚙️ %s ", filepath.Base(d.FilePath))
		if !w.Config.Restart || !FileExists(d.FilePath) {
			err = ioutil.WriteFile(d.FilePath, d.Contents, os.ModePerm)
		}
	default:
		err = errors.New("unrecognized work type")
	}

	if err == nil {
		cli.Logf("⚙️  %s ✅\n", fileName)
	} else {
		cli.Logf("⚙️  %s ☠️\n", fileName)
	}
	return err
}
