package main

import (
	"os"
	"os/signal"
	"syscall"
)

func handleSigterm(stopChan chan struct{}) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM)
	<-signals
	log().Info("Received SIGTERM. Terminating...")
	close(stopChan)
}
