package main

import (
	"fmt"
	"log"
)

// get the latest version of the toolbelt
func latest() {
	latest, err := findLatestVersion()
	if err != nil {
		log.Fatalf("Failed to fetch the latest version: %v\n", err)
	}

	fmt.Printf("The current version is %s\n", latest.Version)
}
