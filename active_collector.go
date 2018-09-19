package main

import "time"

type activeCollector struct {
	kube *Kube
}

func newActiveCollector(kube *Kube) *activeCollector {
	a := &activeCollector{
		kube: kube,
	}
	return a
}

func (a *activeCollector) Run(stop chan struct{}) {
	log().Info("Starting active collector")

	go a.worker()
	<-stop // block until stop closed

	log().Info("Stopping active collector")
	return
}

func (a *activeCollector) worker() {
	for {
		log().Info("active collector tick")
		a.count()
		time.Sleep(time.Second * 5)
	}
}

func (a *activeCollector) count() {
	serviceAccounts, err := a.kube.getServiceAccountList()
	if err != nil {
		log().Errorf("active collector count err: %v", err.Error())
	} else {
		metricActiveTokens.WithLabelValues().Set(float64(len(serviceAccounts.Items)))
	}
}
