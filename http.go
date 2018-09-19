package main

import (
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/gin-gonic/gin"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type HttpServer struct {
	promController *PrometheusController
	kube           *Kube
	config         *Config
}

func newHttpServer(kube *Kube, config *Config) *HttpServer {
	return &HttpServer{
		promController: &PrometheusController{},
		kube:           kube,
		config:         config,
	}
}

func (h *HttpServer) Run() {
	h.promController.registerMetrics()

	router := gin.Default()

	router.GET("/health", h.handlerHealth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// kube config endpoints
	authorized := router.Group("/api/", gin.BasicAuth(*h.config.GinAccounts))
	authorized.GET("/kubeconfig", h.handlerKubeConfig)
	authorized.DELETE("/kubeconfig", h.handlerKubeConfigRevoke)

	// pprof endpoints
	router.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	router.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	router.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	router.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))

	router.Run(fmt.Sprintf(":%d", h.config.Port))
}

func (h *HttpServer) handlerHealth(c *gin.Context) {
	c.String(http.StatusOK, "ok")
}

func (h *HttpServer) handlerKubeConfig(c *gin.Context) {
	username := c.MustGet(gin.AuthUserKey).(string)
	if user, ok := h.config.Users[username]; ok {
		cfg, err := h.kube.getServiceAccountKubeConfig(user.ClusterRole, username)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			log().Error(err.Error())
			metricIssuedTokens.WithLabelValues(username, "error").Inc()
			return
		}
		log().Infof(
			"generated kube config for cluster: (%s) cluster role: (%s) to (%s)",
			h.kube.cluster,
			user.ClusterRole,
			username,
		)
		metricIssuedTokens.WithLabelValues(username, "success").Inc()
		c.String(http.StatusOK, cfg)
		return
	}

	// return default
	c.String(http.StatusInternalServerError, "Ops.. username did not have access configured)")
}

func (h *HttpServer) handlerKubeConfigRevoke(c *gin.Context) {
	username := c.MustGet(gin.AuthUserKey).(string)
	if _, ok := h.config.Users[username]; ok {
		err := h.kube.deleteClusterRoleBindings(username)
		if err != nil {
			c.String(
				http.StatusInternalServerError,
				"Ops.. failed to remove cluster role binding (%s) )", username,
			)
			metricRevokedTokens.WithLabelValues(username, "error").Inc()
			return
		}
		err = h.kube.deleteServiceAccounts(username)
		if err != nil {
			c.String(
				http.StatusInternalServerError,
				"Ops.. failed to remove service account (%s) )", username,
			)
			return
		}
		metricRevokedTokens.WithLabelValues(username, "success").Inc()
		c.String(http.StatusOK, "removed kube config for user (%s)", username)
		return
	}

	// return default
	c.String(http.StatusInternalServerError, "Ops.. username did not have access configured)")
}
