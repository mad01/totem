package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HttpSrv struct {
	router         *gin.Engine
	port           int
	promController *PrometheusController
}

func newHttpSrv(port int) *HttpSrv {
	return &HttpSrv{
		router:         gin.Default(),
		port:           port,
		promController: newPrometheusController(port),
	}
}

func (h *HttpSrv) Run(stopChan chan struct{}) {
	h.router.GET("/health", h.handlerHealth)
	h.router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	h.router.Run(fmt.Sprintf(":%d", h.port))
}

func (h *HttpSrv) handlerHealth(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}
