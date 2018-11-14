package backup

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ushu/udemy-backup/client"
)

var pathSanitizer = strings.NewReplacer("/", "|", ":", " - ")

func getCourseDirectory(cfg *Config, course *client.Course) string {
	slug := getCourseSlug(course)
	return filepath.Join(cfg.RootDir, slug)
}

func getChapterDirectory(cfg *Config, course *client.Course, chapter *client.Chapter) string {
	base := getCourseDirectory(cfg, course)
	if chapter == nil {
		return base // some courses have no chapters
	}

	title := pathSanitizer.Replace(chapter.Title)
	chapterDirName := fmt.Sprintf("%d. %s", chapter.ObjectIndex, title)
	return filepath.Join(base, chapterDirName)
}

func getCourseSlug(course *client.Course) string {
	el := strings.Split(course.URL, "/")
	if len(el) < 2 {
		panic("course has no slug")
	}
	return el[1]
}

func getLecturePrefix(lecture *client.Lecture) string {
	prefix := fmt.Sprintf("%d. %s", lecture.ObjectIndex, lecture.Title)
	return pathSanitizer.Replace(prefix)
}

func getLectureAssetsDirectory(cfg *Config, course *client.Course, lecture *client.Lecture) string {
	dir := getChapterDirectory(cfg, course, lecture.Chapter)
	prefix := getLecturePrefix(lecture)
	return filepath.Join(dir, prefix)
}

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

func linksToFileContents(links []*link) []byte {
	w := new(bytes.Buffer)
	for _, link := range links {
		fmt.Fprintln(w, link.Title)
		fmt.Fprintln(w, link.URL)
		fmt.Fprintln(w)
	}
	return w.Bytes()
}

func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}
