package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/cloud66/cxlogger"
)

var (
	config     Config
	VERSION    string = "dev"
	BUILD_DATE string = ""

	flagDockerPath       string
	flagPollInterval     string
	flagScavengeInterval string
	flagNotifierEndpoint string
	flagCallbackId       string
	flagLog              string
	flagLogLevel         string
)

func main() {
	args := os.Args[1:]

	flag.StringVar(&flagDockerPath, "docker", "/usr/local/bin/docker", "path for docker")
	flag.StringVar(&flagPollInterval, "interval", "5s", "health check intervals in duration (5s, 2m,...)")
	flag.StringVar(&flagScavengeInterval, "scavenge", "10s", "interval to check for missing containers in duration (10s, 5m,...)")
	flag.StringVar(&flagNotifierEndpoint, "notification", "https://app.cloud66.com/containers/status/", "notification endpoint")
	flag.StringVar(&flagCallbackId, "callback", "", "callback id for notification")
	flag.StringVar(&flagLog, "log", "/var/log/cloud66-gocker.log", "log output")
	flag.StringVar(&flagLogLevel, "log-level", "warn", "log level")
	flag.Parse()

	// default logging
	cxlogger.Initialize(flagLog, flagLogLevel)
	cxlogger.Info("Starting")

	if len(args) > 0 && args[0] == "help" {
		flag.PrintDefaults()
		return
	}

	if len(args) > 0 && args[0] == "update" {
		fmt.Println("Updating gocker")
		runUpdate()
		return
	}

	if len(args) > 0 && args[0] == "version" {
		fmt.Printf("gocker version: %s and build date: %s\n", VERSION, BUILD_DATE)
		return
	}

	pollInterval, err := time.ParseDuration(flagPollInterval)
	if err != nil {
		fmt.Printf("Invalid poll interval %s", flagPollInterval)
		os.Exit(-1)
	}

	scavengeInterval, err := time.ParseDuration(flagScavengeInterval)
	if err != nil {
		fmt.Printf("Invalid scavenge interval %s", flagScavengeInterval)
		os.Exit(-1)
	}

	config = Config{
		DockerPath:       flagDockerPath,
		PollInterval:     pollInterval,
		ScavengeInterval: scavengeInterval,
		ScavengeTimeout:  scavengeInterval * 2,
		Notifier:         &Notifier{endpoint: flagNotifierEndpoint},
		CallbackId:       flagCallbackId,
	}

	// we are here, get the automatic updates going
	go autoUpdate()

	manager := Manager{}
	manager.hasLocalState = false

	// do the first started poll before other stuff starts
	// this will set the local state
	manager.performPoll()
	go manager.startPolling()
	go manager.startScavenger()
	go manager.startRefresher()
	select {}
}

func autoUpdate() {
	for _ = range time.Tick(30 * time.Minute) {
		if runUpdate() {
			fmt.Printf("Shutting down so new version can start!")
			os.Exit(0)
		}
	}
}
