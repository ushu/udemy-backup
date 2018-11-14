package pool

import (
	"io"
	"os"

	"github.com/ushu/udemy-backup/client"
)

func downloadURLToFile(c *client.Client, url, filePath string) error {
	tmpPath := filePath + ".tmp"

	// open file for writing
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	// connect to the backend to get the file
	res, err := c.RawGET(url)
	if err != nil {
		f.Close()
		return err
	}

	// load all the data into the local file
	_, err = io.Copy(f, res.Body)
	res.Body.Close()
	if err != nil {
		f.Close()
		return err
	}

	// finally move the temp file into the final place
	err = f.Close()
	if err != nil {
		return err
	}
	return os.Rename(tmpPath, filePath)
}

func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
