package main

import (
	"fmt"
)

var (
	Version = "not set"
)

func getVersion() string {
	return fmt.Sprintf("Version: %v\n", Version)
}
