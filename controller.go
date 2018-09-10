package main

import (
	"time"
)

type controller struct {
	interval time.Duration
	kube     *Kube
	stopChan chan struct{}
	httpSrv  *HttpSrv
}

func newController(kube *Kube, interval time.Duration, port int) *controller {
	c := &controller{
		interval: interval,
		kube:     kube,
		stopChan: make(chan struct{}),
		httpSrv:  newHttpSrv(port, kube),
	}
	return c
}

func (c *controller) Run() {
	log().Info("Starting controller")

	go c.worker(c.stopChan)
	go c.httpSrv.Run(c.stopChan)
	go handleSigterm(c.stopChan)

	<-c.stopChan // block until stopchan closed

	log().Info("Stopping controller")
	return
}
