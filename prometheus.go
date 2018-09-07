package main

// TODO: create init prometheus stuff for monitoring
import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	metricPrefix = "totem"
)

var (
	// active services
	metricActiveServicesEventsCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: getMetricPrefix("active_services_events"),
		Help: "number of service events"},
		[]string{},
	)
)

func getMetricPrefix(name string) string {
	return fmt.Sprintf("%v_%v", metricPrefix, name)
}

type PrometheusController struct {
	port int
	addr string
}

func newPrometheusController(port int) *PrometheusController {
	p := &PrometheusController{
		port: port,
		addr: fmt.Sprintf(":%v", port),
	}
	return p
}

func (p *PrometheusController) registerMetrics() {
	prometheus.MustRegister(metricActiveServicesEventsCounter)
}
