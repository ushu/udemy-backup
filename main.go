package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

const Version = "0.2.0"
const About = `Make backups of Udemy course contents for online usage.`

// Flag values
var (
	showHelp    bool
	showVersion bool
	downloadAll bool
	quiet       bool
)

func init() {
	flag.BoolVar(&showHelp, "help", false, "show usage info")
	flag.BoolVar(&showVersion, "version", false, "show version number")
	flag.BoolVar(&downloadAll, "all", false, "download all the courses enrolled by the user")
	flag.BoolVar(&quiet, "quiet", false, "disable output messages")
	flag.Usage = usage
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("[udemy-backup] ")
	flag.Parse()
	if showHelp {
		usage()
		return
	}
	if showVersion {
		fmt.Printf("%s v%s\n", os.Args[0], Version)
		return
	}
	if quiet {
		log.SetOutput(ioutil.Discard)
	}
}

func usage() {
	fmt.Printf("Usage: %s [COURSE URL|COURSE ID]", os.Args[0])
	fmt.Print("\n\n")
	fmt.Printf(About)
	fmt.Print("\n\n")
	fmt.Println("OPTIONS:")
	flag.PrintDefaults()
}
