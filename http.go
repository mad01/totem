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
	kube           *Kube
}

func newHttpSrv(port int, kube *Kube) *HttpSrv {
	return &HttpSrv{
		router:         gin.Default(),
		port:           port,
		promController: newPrometheusController(port),
		kube:           kube,
	}
}

func (h *HttpSrv) Run() {
	h.router.GET("/health", h.handlerHealth)
	h.router.GET("/metrics", gin.WrapH(promhttp.Handler()))
	h.router.GET("/kubeconfig/:access/:name", h.handlerKubeConfig)
	h.router.Run(fmt.Sprintf(":%d", h.port))
}

func (h *HttpSrv) handlerHealth(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *HttpSrv) handlerKubeConfig(c *gin.Context) {
	accessLevel := c.Param("access")
	name := c.Param("name")
	cfg, err := h.kube.getServiceAccountKubeConfig(accessLevel, name)
	if err != nil {
		c.AbortWithStatus(http.StatusInternalServerError)
		log().Error(err.Error())
	}

	log().Infof(
		"generated kube config for cluster: (%s) with access level: (%s) to (%s)",
		h.kube.cluster,
		accessLevel,
		name,
	)
	c.String(http.StatusOK, cfg)
}

// todo: implement basic auth
