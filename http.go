package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var userAccessLevel = map[string]string{
	"admin":     "admin",
	"alexander": "admin",
	"foo":       "edit",
	"bar":       "view",
}

type HttpServer struct {
	router         *gin.Engine
	promController *PrometheusController
	kube           *Kube
	config         *Config
}

func newHttpServer(kube *Kube, config *Config) *HttpServer {
	return &HttpServer{
		router:         gin.Default(),
		promController: &PrometheusController{},
		kube:           kube,
		config:         config,
	}
}

func (h *HttpServer) Run() {
	h.router.GET("/health", h.handlerHealth)
	h.router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	authorized := h.router.Group("/api/", gin.BasicAuth(*h.config.GinAccounts))
	authorized.GET("/kubeconfig", h.handlerKubeConfig)

	h.router.Run(fmt.Sprintf(":%d", h.config.Port))
}

func (h *HttpServer) handlerHealth(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *HttpSrv) handlerKubeConfig(c *gin.Context) {
	user := c.MustGet(gin.AuthUserKey).(string)
	if accessLevel, ok := userAccessLevel[user]; ok {
		cfg, err := h.kube.getServiceAccountKubeConfig(accessLevel, user)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log().Error(err.Error())
		}
		log().Infof(
			"generated kube config for cluster: (%s) with access level: (%s) to (%s)",
			h.kube.cluster,
			accessLevel,
			user,
		)
		c.String(http.StatusOK, cfg)
	} else {
		c.String(http.StatusInternalServerError, "Ops.. user did not have access configured)")
	}

}
