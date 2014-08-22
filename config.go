package main

import (
	"time"
)

type Config struct {
	DockerPath       string
	PollInterval     time.Duration
	ScavengeInterval time.Duration
	ScavengeTimeout  time.Duration
	Notifier         *Notifier
}
