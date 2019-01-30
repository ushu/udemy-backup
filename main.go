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
	}

	// we use a pull of workers
	nprocs := 4 //runtime.GOMAXPROCS(0)
	var wg sync.WaitGroup
	wg.Add(nprocs)
	ch := make(chan backup.Asset, nprocs)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for i := 0; i < nprocs; i++ {
		go func() {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case a, ok := <-ch:
					if !ok {
						return
					}
					var cerr error
					if a.RemoteURL != "" {
						cerr = downloadURLToFile(client.HTTPClient, a.RemoteURL, a.LocalPath)
					} else if len(a.Contents) > 0 {
						cerr = ioutil.WriteFile(a.LocalPath, a.Contents, os.ModePerm)
					}
					if cerr != nil {
						err = cerr
						cancel()
						return
					}
					if !quiet {
						bar.Increment()
					}
				}
			}
		}()
	}

	// push all the assets
	go func() {
		defer close(ch)
		for _, a := range assets {
			select {
			case <-ctx.Done():
				return
			default:
				ch <- a
			}
		}
	}()

	wg.Wait()
	if !quiet {
		if err != nil {
			bar.FinishPrint("OK")
		} else {
			bar.FinishPrint("KO")
		}
	}
	return err
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
