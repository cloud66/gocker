package main

import (
	"flag"
	)

func main() {
	flag.Parse()

	manager := Manager { }

  go manager.startPolling()
	go manager.startScavenger()
  select {}
}
