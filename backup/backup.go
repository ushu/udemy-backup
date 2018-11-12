package backup

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/spf13/viper"
	"github.com/ushu/udemy-backup/cli"
	"github.com/ushu/udemy-backup/client"
)

var pathSanitizer = strings.NewReplacer("/", "|", ":", " - ")

type work struct {
	Lecture *client.Lecture
	Title   string
	Asset   interface{}
}

type link struct {
	Title string
	URL   string
}

func Run(ctx context.Context, course *client.Course) error {
	var err error
	// contextual values
	c := GetClient(ctx)
	res := viper.GetInt("resolution")
	numWorkers := viper.GetInt("concurrency")
	restart := viper.GetBool("restart")

	// first list all the lectures for the course
	cli.Logf("⚙️  Loading lectures: ")
	lectures, err := c.LoadFullCurriculum(course.ID)
	if err != nil {
		cli.Log()
		return err
	}
	cli.Logf("found %d lectures\n", len(lectures))

	// local "cancel" function
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// start numWorkers workers that comsume incoming lectures
	cli.Logf("⚙️  Stating download with %d workers\n", numWorkers)
	ch := make(chan *work, numWorkers)
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for {
				// pull the ctx for (eventual) cancelation
				select {
				case <-ctx.Done():
					return
				case w, ok := <-ch:
					if !ok {
						return
					}
					dir := getChapterDirectory(course, w.Lecture.Chapter)
					prefix := getLecturePrefix(w.Lecture)
					title := pathSanitizer.Replace(w.Title)

					switch a := w.Asset.(type) {
					case *client.Video:
						// we build the final name for the downloaded file
						fileName := fmt.Sprintf("%s-%s.mp4", prefix, title)
						p := filepath.Join(dir, fileName)

						if restart && FileExists(p) {
							cli.Log("💡  skipping existing file:", fileName)
							continue
						}

						// and download it !
						e := downloadURLToFile(c, a.File, p)
						if e != nil {
							err = e
							cancel()
							return
						}
						cli.Logf("🎬 %s ✅\n", fileName)
					case *client.File:
						dir := getLectureAssetsDirectory(course, w.Lecture)
						p := filepath.Join(dir, title)

						if restart && FileExists(p) {
							cli.Log("💡  skipping existing file:", filepath.Join(prefix, title))
							continue
						}

						// and download it !
						e := downloadURLToFile(c, a.File, p)
						if e != nil {
							err = e
							cancel()
							return
						}
						cli.Logf("🔗 %s ✅\n", filepath.Join(prefix, title))
					case []*link:
						dir := getLectureAssetsDirectory(course, w.Lecture)
						fileName := fmt.Sprintf("%s.txt", title)
						p := filepath.Join(dir, fileName)

						if restart && FileExists(p) {
							cli.Log("💡  skipping existing file:", filepath.Join(prefix, fileName))
							continue
						}

						e := dumpLinks(p, a)
						if e != nil {
							err = e
							cancel()
							return
						}
						cli.Logf("🔗  %s ✅\n", filepath.Join(prefix, fileName))
					case *client.Caption:
						base := strings.TrimSuffix(title, filepath.Ext(title))
						ext := filepath.Ext(a.FileName)
						locale := a.Locale.Locale
						fileName := fmt.Sprintf("%s-%s.%s%s", prefix, base, locale, ext)
						p := filepath.Join(dir, fileName)

						if restart && FileExists(p) {
							cli.Log("💡  skipping existing file:", fileName)
							continue
						}

						e := downloadURLToFile(c, a.URL, p)
						if e != nil {
							err = e
							cancel()
							return
						}
						cli.Logf("🎙  %s ✅\n", fileName)
					}
				}
			}
		}()
	}

	// push all the lectures for processing, and download complete info
	e := buildCourseDirectory(course)
	// make the path
	if e != nil {
		err = e
		cancel()
	} else {
		var currentChapter *client.Chapter
	Loop:
		for _, l := range lectures {
			select {
			case <-ctx.Done():
				break
			default:
			}

			if l.Chapter != nil && currentChapter != l.Chapter {
				if e = buildChapterDirectory(course, l.Chapter); e != nil {
					err = e
					cancel()
					break Loop
				}
			}

			// enqueue all the videos for downloading
			var videos []*client.Video
			if l.Asset.DownloadUrls != nil {
				videos = l.Asset.DownloadUrls.Video
			} else if l.Asset.StreamUrls != nil {
				videos = l.Asset.StreamUrls.Video
			}
			if len(videos) > 0 {
				// filter using the preferred resolution
				video := filterVideos(res, videos)
				if video != nil {
					// and finally trigger the workers
					select {
					case ch <- &work{
						Lecture: l,
						Title:   video.Label,
						Asset:   video,
					}:
					case <-ctx.Done():
						break Loop
					}
				}

				for _, c := range l.Asset.Captions {
					select {
					case ch <- &work{
						Lecture: l,
						Title:   video.Label,
						Asset:   c,
					}:
					case <-ctx.Done():
						break Loop
					}
				}
			}

			// also download additional files
			if len(l.SupplementatyAssets) > 0 {
				if e = buildLectureAssetsDirectory(course, l); e != nil {
					err = e
					cancel()
					break Loop
				}

				var links []*link
				for _, a := range l.SupplementatyAssets {
					if a.AssetType == "ExternalLink" {
						links = append(links, &link{
							Title: a.Title,
							URL:   a.ExternalURL,
						})
					}
					if a.DownloadUrls == nil {
						continue
					}
					for _, f := range a.DownloadUrls.File {
						select {
						case ch <- &work{
							Lecture: l,
							Title:   a.Title,
							Asset:   f,
						}:
						case <-ctx.Done():
							break Loop
						}
					}
				}
				if len(links) > 0 {
					select {
					case ch <- &work{
						Lecture: l,
						Title:   "links",
						Asset:   links,
					}:
					case <-ctx.Done():
						break Loop
					}
				}
			}
		}
		close(ch)
	}

	wg.Wait()
	return err
}

func filterVideos(res int, videos []*client.Video) *client.Video {
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
		if res > 0 && vres == res {
			// perfect match
			return v
		}
		// find the highest resolution available
		if vres > currentRes {
			video = v
			currentRes = vres
		}
	}
	return video
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

func FileExists(name string) bool {
	_, err := os.Stat(name)
	return !os.IsNotExist(err)
}

func dumpLinks(filePath string, links []*link) error {
	// create of replace file
	f, err := os.Create(filePath)
	if err != nil && !os.IsExist(err) {
		return err
	}
	defer f.Close()

	// use buffered IO for speed
	w := bufio.NewWriter(f)
	defer w.Flush()

	// and we dump all available data
	for _, link := range links {
		_, err = fmt.Fprintln(w, link.Title)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w, link.URL)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(w)
		if err != nil {
			return err
		}
	}
	return nil
}

func buildCourseDirectory(course *client.Course) error {
	p := getCourseDirectory(course)
	err := os.MkdirAll(p, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func getCourseDirectory(course *client.Course) string {
	dir := viper.GetString("dir")
	slug := getCourseSlug(course)
	return filepath.Join(dir, slug)
}

func getCourseSlug(course *client.Course) string {
	return strings.Split(course.URL, "/")[1]
}

func buildLectureAssetsDirectory(course *client.Course, lecture *client.Lecture) error {
	p := getLectureAssetsDirectory(course, lecture)
	err := os.MkdirAll(p, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func getLectureAssetsDirectory(course *client.Course, lecture *client.Lecture) string {
	dir := getChapterDirectory(course, lecture.Chapter)
	prefix := getLecturePrefix(lecture)
	return filepath.Join(dir, prefix)
}

func buildChapterDirectory(course *client.Course, chapter *client.Chapter) error {
	p := getChapterDirectory(course, chapter)
	err := os.MkdirAll(p, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func getChapterDirectory(course *client.Course, chapter *client.Chapter) string {
	dir := viper.GetString("dir")
	slug := getCourseSlug(course)
	var chapterPath string
	if chapter != nil {
		title := pathSanitizer.Replace(chapter.Title)
		chapterPath = fmt.Sprintf("%d. %s", chapter.ObjectIndex, title)
	}
	return filepath.Join(dir, slug, chapterPath)
}

func getLecturePrefix(lecture *client.Lecture) string {
	prefix := fmt.Sprintf("%d. %s", lecture.ObjectIndex, lecture.Title)
	return pathSanitizer.Replace(prefix)
}
