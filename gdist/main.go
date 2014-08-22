package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/kr/s3"
)

var subcmds = map[string]func(){
	"build":   build,
	"publish": publish,
	"latest":  latest,
}

const (
	sourceGit = "git@github.com:cloud66/gocker.git"
)

var (
	buildDir   string
	publishDir string
	s3keys     = s3.Keys{
		AccessKey: os.Getenv("AWS_ACCESS_KEY"),
		SecretKey: os.Getenv("AWS_SECRET_KEY"),
	}
	flags       flag.FlagSet
	flagVersion string
	flagBranch  string
)

func usage() {
	fmt.Fprintln(os.Stderr, "Usage: gdist (build|publish|latest) -v <version> -b <branch>")
	os.Exit(2)
}

func main() {
	log.SetFlags(log.Lshortfile)

	flags.StringVar(&flagVersion, "v", "dev", "build version")
	flags.StringVar(&flagBranch, "b", "master", "build branch")

	if len(os.Args) == 1 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	args := os.Args[2:]

	if err := flags.Parse(args); err != nil {
		os.Exit(2)
	}

	if cmd != "build" && cmd != "publish" && cmd != "latest" {
		usage()
	}

	pwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
		os.Exit(2)
	}
	buildDir = filepath.Join(pwd, "tmp")
	publishDir = filepath.Join(pwd, "build")

	f := subcmds[cmd]
	if f == nil {
		usage()
	}

	f()
}
