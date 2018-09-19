package main

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricPrefix = "totem"
)

var (
	metricIssuedTokens = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: getMetricPrefix("issued_configs"),
		Help: "number issued kube configs"},
		[]string{"username", "status"},
	)

	metricRevokedTokens = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: getMetricPrefix("revoked_configs"),
		Help: "number revoked kube configs"},
		[]string{"username", "status"},
	)
)

func getMetricPrefix(name string) string {
	return fmt.Sprintf("%v_%v", metricPrefix, name)
}

type PrometheusController struct {
}

func (p *PrometheusController) registerMetrics() {
	prometheus.MustRegister(metricIssuedTokens)
	prometheus.MustRegister(metricRevokedTokens)
}
