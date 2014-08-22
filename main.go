package main

import ("flag"
				"time"
				)

var config Config

func main() {
	flag.Parse()

	// TODO: load from flags
	config = Config {
		DockerPath: "/usr/local/bin/docker",
		PollInterval : 2 * time.Second,
		ScavengeInterval: 10 * time.Second,
		ScavengeTimeout: 20 * time.Second,
	}

	manager := Manager { }

  go manager.startPolling()
	go manager.startScavenger()
  select {}
}
