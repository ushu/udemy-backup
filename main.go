package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"

	"github.com/ushu/udemy-backup/backup"
	"github.com/ushu/udemy-backup/client"
	"github.com/ushu/udemy-backup/client/lister"
	pb "gopkg.in/cheggaaa/pb.v1"
)

// Version of the tool
var Version = "0.2.0"

// Help message (before options)
const usageDescription = `Usage: udemy-backup

Make backups of Udemy course contents for offline usage.

OPTIONS:
`

// Flag values
var (
	showHelp    bool
	showVersion bool
	downloadAll bool
	quiet       bool
)

// Number of parallel workers
var concurrency = 4

func init() {
	flag.BoolVar(&downloadAll, "a", false, "download all the courses enrolled by the user")
	flag.BoolVar(&showHelp, "h", false, "show usage info")
	flag.BoolVar(&quiet, "q", false, "disable output messages")
	flag.BoolVar(&showVersion, "v", false, "show version number")
	flag.Usage = func() {
		fmt.Print(usageDescription)
		flag.PrintDefaults()
	}
	log.SetFlags(0)
	log.SetPrefix("")
}

func main() {
	flag.Parse()
	ctx := context.Background()

	// Parse flags
	if showHelp {
		flag.Usage()
		return
	}
	if showVersion {
		fmt.Printf("v%s\n", Version)
		return
	}
	if quiet {
		log.SetOutput(ioutil.Discard)
	}

	// Connect to the Udemy backend
	e, p, err := askCredentials()
	if err != nil {
		log.Fatal(err)
	}
	c := client.New()
	_, err = c.Login(ctx, e, p)
	if err != nil {
		log.Fatal(err)
	}

	// list all the courses
	l := lister.New(c)
	courses, err := l.ListAllCourses(ctx)
	if err != nil {
		log.Fatal(err)
	}

	// we're logged in !
	if downloadAll {
		for _, course := range courses {
			log.Printf("🚀 %s", course.Title)
			if err = downloadCourse(ctx, c, course); err != nil {
				log.Fatal(err)
			}
		}
	} else {
		course, err := selectCourse(courses)
		if err != nil {
			log.Fatal(err)
		}
		if err = downloadCourse(ctx, c, course); err != nil {
			log.Fatal(err)
		}
	}
}

func downloadCourse(ctx context.Context, client *client.Client, course *client.Course) error {
	var err error

	// list all the available course elements
	b := backup.New(client, ".", false)
	assets, dirs, err := b.ListCourseAssets(ctx, course)
	if err != nil {
		return err
	}

	// create all the required directories
	for _, d := range dirs {
		os.MkdirAll(d, 0755)
	}

	// start the bar
	var bar *pb.ProgressBar
	if !quiet {
		bar = pb.StartNew(len(assets))
		defer bar.FinishPrint("")
	}

	// start a cancelable context
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// we use a pull of workers
	chwork := make(chan backup.Asset)      // assets to process get enqueued here
	cherr := make(chan error, concurrency) // download results

	// start the workers
	var wg sync.WaitGroup
	wg.Add(concurrency)
	for i := 0; i < concurrency; i++ {
		go func() {
			defer wg.Done()
			for a := range chwork {
				if a.RemoteURL != "" {
					cherr <- downloadURLToFile(client.HTTPClient, a.RemoteURL, a.LocalPath)
				} else if len(a.Contents) > 0 {
					cherr <- ioutil.WriteFile(a.LocalPath, a.Contents, os.ModePerm)
				}
				if !quiet {
					bar.Increment()
				}
			}
		}()
	}

	// and the "pusher" goroutine
	go func() {
		// enqueue all assets (unless we cancel)
		for _, a := range assets {
			select {
			case <-ctx.Done():
				break
			case chwork <- a:
			}
		}
		// we close channels on "enqueing" side to avoid panics
		close(chwork) // <- will stop the workers
		wg.Wait()
		close(cherr) // <- we close when we are sure there won't be a "write"
	}()

	// we wait for an error (if any)
	for err := range cherr {
		if err != nil {
			return err // <- will cancel the context, then the "pusher", then the workers
		}
	}
	return nil
}

func downloadURLToFile(c *http.Client, url, filePath string) error {
	tmpPath := filePath + ".tmp"

	// open file for writing
	f, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	// connect to the backend to get the file
	res, err := c.Get(url)
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
