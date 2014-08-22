package main

import (
	"log"
	"os"
	"os/exec"
	"strings"
)

func build() {
	// checkout code
	if err := cloneRepo(sourceGit, flagBranch, buildDir); err != nil {
		log.Fatal(err)
	}

	// build it
	if err := buildApp(buildDir, flagVersion); err != nil {
		log.Fatal(err)
	}
}

func cloneRepo(repo, branch, dir string) error {
	if err := os.RemoveAll(dir); err != nil {
		return err
	}
	if _, err := cmd("git", "clone", "-b", branch, repo, dir); err != nil {
		return err
	}
	return nil
}

func buildApp(dir, ver string) error {
	if _, err := cmd("goxc", "-include=''", "-bc=linux,windows,darwin", "-pv="+ver, "-d="+publishDir, "-main-dirs-exclude=gockerdist", "-n=gocker", "-wd="+dir); err != nil {
		return err
	}
	return nil
}

func cmd(arg ...string) ([]byte, error) {
	log.Println(strings.Join(arg, " "))
	cmd := exec.Command(arg[0], arg[1:]...)
	cmd.Stderr = os.Stderr
	return cmd.Output()
}
