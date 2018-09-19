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

	metricRevokedHTTPTokens = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: getMetricPrefix("revoked_configs"),
		Help: "number revoked kube configs"},
		[]string{"username", "status"},
	)
	metricRevokedCleanupTokens = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: getMetricPrefix("revoked_cleanup_configs"),
		Help: "number revoked kube configs"},
		[]string{"status", "kind"},
	)

	metricActiveTokens = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: getMetricPrefix("currently_active_kube_configs"),
		Help: "number active kube configs"},
		[]string{},
	)
)

func getMetricPrefix(name string) string {
	return fmt.Sprintf("%v_%v", metricPrefix, name)
}

type PrometheusController struct {
}

func (p *PrometheusController) registerMetrics() {
	prometheus.MustRegister(metricIssuedTokens)
	prometheus.MustRegister(metricRevokedHTTPTokens)
	prometheus.MustRegister(metricRevokedCleanupTokens)
	prometheus.MustRegister(metricActiveTokens)
}
