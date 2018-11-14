package backup

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ushu/udemy-backup/backup/config"
	"github.com/ushu/udemy-backup/client"
)

var pathSanitizer = strings.NewReplacer("/", "|", ":", " - ")

func getCourseDirectory(cfg *config.Config, course *client.Course) string {
	slug := getCourseSlug(course)
	return filepath.Join(cfg.RootDir, slug)
}

func getChapterDirectory(cfg *config.Config, course *client.Course, chapter *client.Chapter) string {
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

func getLectureAssetsDirectory(cfg *config.Config, course *client.Course, lecture *client.Lecture) string {
	dir := getChapterDirectory(cfg, course, lecture.Chapter)
	prefix := getLecturePrefix(lecture)
	return filepath.Join(dir, prefix)
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
