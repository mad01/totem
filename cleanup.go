package main

import (
	"time"
)

type cleanupController struct {
	interval      time.Duration
	tokenLifetime time.Duration
	kube          *Kube
	stopChan      chan struct{}
}

func newCleanupController(kube *Kube, interval, lifetime time.Duration) *cleanupController {
	c := &cleanupController{
		interval:      interval,
		tokenLifetime: lifetime,
		kube:          kube,
		stopChan:      make(chan struct{}),
	}
	return c
}

func (c *cleanupController) Run() {
	log().Info("Starting cleanup controller")

	go c.worker(c.stopChan)
	<-c.stopChan // block until stopchan closed

	log().Info("Stopping cleanup controller")
	return
}

func (c *cleanupController) worker(stopChan chan struct{}) {
	for {
		log().Info("cleanup tick")
		c.deleteTimedOutClusterRoleBindings()
		c.deleteTimedOutServiceAccounts()
		time.Sleep(c.interval)
	}
}

func (c *cleanupController) deleteTimedOutServiceAccounts() {
	serviceAccounts, err := c.kube.getServiceAccountList()
	if err != nil {
		log().Error(err)
	}
	for _, sa := range serviceAccounts.Items {
		if createdAtString, ok := sa.Annotations[annotationCreatedAt]; ok {
			createdAt, err := time.Parse(timeFormat, createdAtString)
			if err != nil {
				log().Errorf("parsing annotation of service account (%s): %v", sa.Name, err)
				continue
			}

			if !inTimeSpan(createdAt, c.tokenLifetime) {
				err = c.kube.deleteServiceAccount(sa.Name)
				if err != nil {
					log().Errorf("deleting service account (%s): %v", sa.Name, err)
					continue
				}
				log().Infof("service account (%s) outside time span, deleting it", sa.Name)
				metricRevokedHTTPTokens.WithLabelValues(username, "success", "cleanup").Inc()
			} else if inTimeSpan(createdAt, c.tokenLifetime) {
				log().Infof("service account (%s) still in time span ", sa.Name)
			}

		}
	}

}

func (c *cleanupController) deleteTimedOutClusterRoleBindings() {
	clusterRoleBindings, err := c.kube.getClusterRoleBindingList()
	if err != nil {
		log().Error(err)
	}
	for _, crb := range clusterRoleBindings.Items {
		if createdAtString, ok := crb.Annotations[annotationCreatedAt]; ok {
			createdAt, err := time.Parse(timeFormat, createdAtString)
			if err != nil {
				log().Errorf("parsing of cluster role binding (%s): %v", crb.Name, err)
				continue
			}

			if !inTimeSpan(createdAt, c.tokenLifetime) {
				err = c.kube.deleteClusterRoleBinding(crb.Name)
				if err != nil {
					log().Errorf("deleting of cluster role binding (%s): %v", crb.Name, err)
					continue
				}
				log().Infof("cluster role binding (%s) outside time span, deleting it ", crb.Name)
			} else if inTimeSpan(createdAt, c.tokenLifetime) {
				log().Infof("cluster role binding (%s) still in time span ", crb.Name)
			}

		}
	}

}

func inTimeSpan(start time.Time, lifetime time.Duration) bool {
	end := start.Local().Add(lifetime)
	now := time.Now()

	return now.After(start) && now.Before(end)
}
