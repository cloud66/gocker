package main

import (
	"os/exec"
	"time"

	"github.com/cloud66/cxlogger"
)

// this is a single docker process that has been observed at least once
type DockerProcess struct {
	uid            string
	lastObservedAt time.Time
}

func (dockerProcess *DockerProcess) Inspect() (string, error) {
	cxlogger.Log.Info("Getting inspect meta data")
	cmd := exec.Command(config.DockerPath, "inspect", dockerProcess.uid)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
