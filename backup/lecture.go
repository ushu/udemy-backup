package backup

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ushu/udemy-backup/client"
)

type link struct {
	Title string
	URL   string
}

func BackupLecture(ctx context.Context, cfg *Config, course *client.Course, lecture *client.Lecture) error {
	prefix := getLecturePrefix(lecture)

	// grab the pool
	workerPool, ok := FromContext(ctx)
	if !ok {
		return errors.New("could not locate the worker pool")
	}

	// flag for building the (optional) assets dir
	var assetsDirectoryBuilt bool

	// now we traverse the Lecture struct, and enqueue all the necessary work
	// first the video stream, if any
	videos := findVideos(lecture)
	video := filterVideos(videos, cfg.Resolution)
	if video != nil {
		dir := getChapterDirectory(cfg, course, lecture.Chapter)
		path := filepath.Join(dir, prefix+".mp4")
		url := video.File
		if err := workerPool.EnqueueDowload(ctx, cfg, url, path); err != nil {
			return err
		}

		// when the stream is found, we also look up the captions
		if cfg.LoadSubtitles && lecture.Asset != nil && len(lecture.Asset.Captions) > 0 {
			if !assetsDirectoryBuilt {
				if err := buildLectureAssetsDirectory(cfg, course, lecture); err != nil {
					return err
				}
				assetsDirectoryBuilt = true
			}
			for _, c := range lecture.Asset.Captions {
				dir := getLectureAssetsDirectory(cfg, course, lecture)
				ext := filepath.Ext(c.FileName)
				locale := c.Locale.Locale
				fileName := fmt.Sprintf("%s.%s%s", prefix, locale, ext)
				path := filepath.Join(dir, fileName)
				if err := workerPool.EnqueueDowload(ctx, cfg, c.URL, path); err != nil {
					return err
				}
			}
		}
	}

	// additional files
	if len(lecture.SupplementatyAssets) > 0 {
		if !assetsDirectoryBuilt {
			if err := buildLectureAssetsDirectory(cfg, course, lecture); err != nil {
				return err
			}
		}
		dir := getLectureAssetsDirectory(cfg, course, lecture)

		var links []*link
		for _, a := range lecture.SupplementatyAssets {
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
				path := filepath.Join(dir, f.Label)
				url := f.File
				if err := workerPool.EnqueueDowload(ctx, cfg, url, path); err != nil {
					return err
				}

			}
		}
		// finally, if we found one or more links, we create a "links.txt" file
		if len(links) > 0 {
			path := filepath.Join(dir, "links.txt")
			contents := linksToFileContents(links)
			if err := workerPool.EnqueueWrite(ctx, cfg, path, contents); err != nil {
				return err
			}
		}
	}

	return nil
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

func buildLectureAssetsDirectory(cfg *Config, course *client.Course, lecture *client.Lecture) error {
	p := getLectureAssetsDirectory(cfg, course, lecture)
	return os.MkdirAll(p, 0755)
}
