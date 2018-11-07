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
	dir := viper.GetString("dir")
	restart := viper.GetBool("restart")

	// first list all the lectures for the course
	cli.Logf("⚙️  Loading lectures: ")
	lectures, err := c.ListAllLectures(course.ID, true)
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
	slug := getCourseSlug(course)
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
					prefix := getLecturePrefix(w.Lecture)
					title := pathSanitizer.Replace(w.Title)

					switch a := w.Asset.(type) {
					case *client.Video:
						// we build the final name for the downloaded file
						fileName := fmt.Sprintf("%s-%s.mp4", prefix, title)
						p := filepath.Join(dir, slug, fileName)

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
						// we build the final name for the downloaded file
						p := filepath.Join(dir, slug, prefix, title)

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
						cli.Logf("🔗  %s ✅\n", filepath.Join(prefix, title))
					case []*link:
						fileName := fmt.Sprintf("%s.txt", title)
						p := filepath.Join(dir, slug, prefix, fileName)

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
						p := filepath.Join(dir, slug, fileName)

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
	Loop:
		for _, l := range lectures {
			select {
			case <-ctx.Done():
				break
			default:
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
					case <-ctx.Done():
						break Loop
					case ch <- &work{
						Lecture: l,
						Title:   video.Label,
						Asset:   video,
					}:
					}
				}

				for _, c := range l.Asset.Captions {
					select {
					case <-ctx.Done():
						break Loop
					case ch <- &work{
						Lecture: l,
						Title:   video.Label,
						Asset:   c,
					}:
					}
				}
			}

			// also download additional files
			if len(l.SupplementatyAssets) > 0 {
				e := buildLecureAssetsDirectory(course, l)
				if e != nil {
					err = e
					cancel()
					break
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
						case <-ctx.Done():
							break Loop
						case ch <- &work{
							Lecture: l,
							Title:   a.Title,
							Asset:   f,
						}:
						}
					}
				}
				if len(links) > 0 {
					select {
					case <-ctx.Done():
						break Loop
					case ch <- &work{
						Lecture: l,
						Title:   "links",
						Asset:   links,
					}:
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
	dir := viper.GetString("dir")
	slug := getCourseSlug(course)
	p := filepath.Join(dir, slug)
	err := os.MkdirAll(p, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func buildLecureAssetsDirectory(course *client.Course, lecture *client.Lecture) error {
	dir := viper.GetString("dir")
	slug := getCourseSlug(course)
	prefix := getLecturePrefix(lecture)
	p := filepath.Join(dir, slug, prefix)
	err := os.MkdirAll(p, 0755)
	if err != nil && !os.IsExist(err) {
		return err
	}
	return nil
}

func getCourseSlug(course *client.Course) string {
	return strings.Split(course.URL, "/")[1]
}

func getLecturePrefix(lecture *client.Lecture) string {
	prefix := fmt.Sprintf("%03d-%s", lecture.ObjectIndex, lecture.Title)
	return pathSanitizer.Replace(prefix)
}
