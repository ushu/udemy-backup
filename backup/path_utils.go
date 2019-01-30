package backup

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ushu/udemy-backup/client"
)

var pathSanitizer = strings.NewReplacer("/", "|", ":", " - ")

func getCourseDirectory(rootDir string, course *client.Course) string {
	slug := getCourseSlug(course)
	return filepath.Join(rootDir, slug)
}

func getChapterDirectory(rootDir string, course *client.Course, chapter *client.Chapter) string {
	base := getCourseDirectory(rootDir, course)
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

func getLectureAssetsDirectory(rootDir string, course *client.Course, lecture *client.Lecture) string {
	dir := getChapterDirectory(rootDir, course, lecture.Chapter)
	prefix := getLecturePrefix(lecture)
	return filepath.Join(dir, prefix)
}
