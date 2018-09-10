package main

import (
	"time"
)

type controller struct {
	interval          time.Duration
	tokenLifetime     time.Duration
	kube              *Kube
	stopChan          chan struct{}
	httpServer        *HttpServer
	cleanupController *cleanupController
}

func newController(kube *Kube, interval, lifetime time.Duration, config *Config) *controller {
	c := &controller{
		kube:              kube,
		stopChan:          make(chan struct{}),
		httpServer:        newHttpServer(kube, config),
		cleanupController: newCleanupController(kube, interval, lifetime),
	}
	return c
}

func (c *controller) Run() {
	log().Info("Starting controller")

	go c.httpServer.Run()
	go c.cleanupController.Run()

	go handleSigterm(c.stopChan)

	<-c.stopChan // block until stopchan closed

	log().Info("Stopping controller")
	return
}
