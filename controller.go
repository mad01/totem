package main

import (
	"time"
)

type controller struct {
	interval          time.Duration
	tokenLifetime     time.Duration
	kube              *Kube
	stopChan          chan struct{}
	httpSrv           *HttpSrv
	cleanupController *cleanupController
}

func newController(kube *Kube, interval, lifetime time.Duration, port int) *controller {
	c := &controller{
		kube:              kube,
		stopChan:          make(chan struct{}),
		httpSrv:           newHttpSrv(port, kube),
		cleanupController: newCleanupController(kube, interval, lifetime),
	}
	return c
}

func (c *controller) Run() {
	log().Info("Starting controller")

	go c.httpSrv.Run(c.stopChan)
	go c.cleanupController.Run()

	go handleSigterm(c.stopChan)

	<-c.stopChan // block until stopchan closed

	log().Info("Stopping controller")
	return
}
