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
	activeCollector   *activeCollector
}

func newController(kube *Kube, interval, lifetime time.Duration, config *Config) *controller {
	c := &controller{
		kube:              kube,
		stopChan:          make(chan struct{}),
		httpServer:        newHttpServer(kube, config),
		cleanupController: newCleanupController(kube, interval, lifetime),
		activeCollector:   newActiveCollector(kube),
	}
	return c
}

func (c *controller) Run() {
	log().Info("Starting controller")

	go handleSigterm(c.stopChan)
	go c.cleanupController.Run(c.stopChan)
	go c.activeCollector.Run(c.stopChan)
	go c.httpServer.Run()

	<-c.stopChan // block until stopChan closed

	log().Info("Stopping controller")
	return
}
