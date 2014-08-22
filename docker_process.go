package main

import (
	"time"
	)

// this is a single docker process that has been observed at least once
type DockerProcess struct {
	uid							string
	lastObservedAt	time.Time
}
