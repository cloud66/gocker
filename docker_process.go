package main

import (
	"github.com/golang/glog"
	"os/exec"
	"time"
)

// this is a single docker process that has been observed at least once
type DockerProcess struct {
	uid            string
	lastObservedAt time.Time
}

func (dockerProcess *DockerProcess) Inspect() (string, error) {
	glog.V(5).Info("Getting inspect meta data")
	cmd := exec.Command(config.DockerPath, "inspect", dockerProcess.uid)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
