package backup

//	return os.MkdirAll(p, 0755)

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ushu/udemy-backup/client"
	"github.com/ushu/udemy-backup/client/lister"
)

type Backuper struct {
	Client        *client.Client
	RootDir       string
	LoadSubtitles bool
}

type Asset struct {
	LocalPath string
	RemoteURL string
	Contents  []byte
}

type link struct {
	Title string
	URL   string
}

func New(client *client.Client, rootDir string, loadSubtitles bool) *Backuper {
	return &Backuper{client, rootDir, loadSubtitles}
}

func (b *Backuper) ListCourseAssets(ctx context.Context, course *client.Course) ([]Asset, []string, error) {
	var directories []string
	var assets []Asset

	// then we list all the lectures for the course
	lst := lister.New(b.Client)
	lectures, err := lst.LoadFullCurriculum(ctx, course.ID)
	if err != nil {
		return assets, directories, err
	}
	// we start by creating the necessary directories to hold all the lectures the root dir
	courseDir := getCourseDirectory(b.RootDir, course)
	if !dirExists(courseDir) {
		directories = append(directories, courseDir)
	}

	// now we parse the curriculum
	for _, l := range lectures {
		if chap, ok := l.(*client.Chapter); ok {
			chapDir := getChapterDirectory(b.RootDir, course, chap)
			if !dirExists(chapDir) {
				directories = append(directories, chapDir)
			}
		} else if lecture, ok := l.(*client.Lecture); ok {
			courseAssets, courseDirs := b.ListLectureAssets(course, lecture)
			if err != nil {
				return assets, directories, err
			}
			assets = append(assets, courseAssets...)
			for _, courseDir := range courseDirs {
				if dirExists(courseDir) {
					directories = append(directories, courseDir)
				}
			}
		}
	}

	return assets, directories, nil
}

func (b *Backuper) ListLectureAssets(course *client.Course, lecture *client.Lecture) ([]Asset, []string) {
	var directories []string
	var assets []Asset
	prefix := getLecturePrefix(lecture)

	// flag for building the (optional) assets dir
	assetsDir := getLectureAssetsDirectory(b.RootDir, course, lecture)
	assetsDirectoryBuilt := dirExists(assetsDir)

	// now we traverse the Lecture struct, and enqueue all the necessary work
	// first the video stream, if any
	videos := findVideos(lecture)
	video := filterVideos(videos, 1080)
	if video != nil {
		// enqueue download of the video
		dir := getChapterDirectory(b.RootDir, course, lecture.Chapter)
		assets = append(assets, Asset{
			LocalPath: filepath.Join(dir, prefix+".mp4"),
			RemoteURL: video.File,
		})

		// when the stream is found, we also look up the captions
		if b.LoadSubtitles && lecture.Asset != nil && len(lecture.Asset.Captions) > 0 {
			if !assetsDirectoryBuilt {
				directories = append(directories, assetsDir)
				assetsDirectoryBuilt = true
			}
			for _, c := range lecture.Asset.Captions {
				ext := filepath.Ext(c.FileName)
				locale := c.Locale.Locale
				fileName := fmt.Sprintf("%s.%s%s", prefix, locale, ext)
				assets = append(assets, Asset{
					LocalPath: filepath.Join(dir, fileName),
					RemoteURL: c.URL,
				})
			}
		}
	}

	//
	// additional files
	//
	if len(lecture.SupplementaryAssets) > 0 {
		dir := getLectureAssetsDirectory(b.RootDir, course, lecture)
		if !assetsDirectoryBuilt {
			directories = append(directories, assetsDir)
			assetsDirectoryBuilt = true
		}

		var links []*link
		for _, a := range lecture.SupplementaryAssets {
			// we handle links differently
			if a.AssetType == "ExternalLink" {
				links = append(links, &link{
					Title: a.Title,
					URL:   a.ExternalURL,
				})
				continue
			}
			// and we also assets there is something to download
			if a.DownloadUrls == nil {
				continue
			}
			// now we grab the file, into the assets directory
			for _, f := range a.DownloadUrls.File {
				assets = append(assets, Asset{
					LocalPath: filepath.Join(assetsDir, f.Label),
					RemoteURL: f.File,
				})
			}
		}
		// finally, if we found one or more links, we create a "links.txt" file
		if len(links) > 0 {
			contents := linksToFileContents(links)
			assets = append(assets, Asset{
				LocalPath: filepath.Join(dir, "links.txt"),
				Contents:  contents,
			})
		}
	}

	return assets, directories
}

func findVideos(lecture *client.Lecture) []*client.Video {
	if lecture.Asset.DownloadUrls != nil {
		return lecture.Asset.DownloadUrls.Video
	} else if lecture.Asset.StreamUrls != nil {
		return lecture.Asset.StreamUrls.Video
	}
	return nil
}

func filterVideos(videos []*client.Video, resolution int) *client.Video {
	var video *client.Video
	currentRes := 0
	for _, v := range videos {
		if v.Type != "video/mp4" {
			continue
		}
		vres, err := strconv.Atoi(v.Label)
		if err != nil {
			continue
		}
		if resolution > 0 && vres == resolution {
			// perfect match
			return v
		}
		// or find the highest resolution available
		if vres > currentRes {
			video = v
			currentRes = vres
		}
	}
	return video
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

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func dirExists(name string) bool {
	s, err := os.Stat(name)
	return !os.IsNotExist(err) && s.IsDir()
}
