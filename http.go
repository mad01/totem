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
	authorized.DELETE("/revoke/:name", h.handlerKubeConfigRevokeADMIN)

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
		cfg, err := h.kube.getServiceAccountKubeConfig(&user)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			log().Error(err.Error())
			metricIssuedTokens.WithLabelValues(user.Name, "error").Inc()
			return
		}
		log().Infof(
			"generated kube config for cluster: (%s) cluster role: (%s) to (%s)",
			h.kube.cluster,
			user.ClusterRole,
			user.Name,
		)
		metricIssuedTokens.WithLabelValues(user.Name, "success").Inc()
		c.String(http.StatusOK, cfg)
		return
	}

	// return default
	c.String(http.StatusInternalServerError, "Ops.. username did not have access configured)")
}

func (h *HttpServer) handlerKubeConfigRevoke(c *gin.Context) {
	username := c.MustGet(gin.AuthUserKey).(string)
	if user, ok := h.config.Users[username]; ok {
		err := h.kube.delete(&user)
		if err != nil {
			c.String(http.StatusInternalServerError, err.Error())
			metricRevokedHTTPTokens.WithLabelValues(user.Name, "error").Inc()
			return
		} else {
			metricRevokedHTTPTokens.WithLabelValues(user.Name, "success").Inc()
			c.String(http.StatusOK, "removed kube config for user (%s)", user.Name)
			return
		}
	}

	// return default
	c.String(http.StatusInternalServerError, "Ops.. username did not have access configured)")
}

func (h *HttpServer) handlerKubeConfigRevokeADMIN(c *gin.Context) {
	// allows users with admin to revoke others token
	username := c.MustGet(gin.AuthUserKey).(string)
	userToRemove := c.Param("name")
	if user, ok := h.config.Users[username]; ok {
		if user.isAdmin() {
			if userToRemove == "" {
				c.String(http.StatusInternalServerError, "missing username param /api/revoke/:name:")
				metricRevokedHTTPTokensADMIN.WithLabelValues(username, "error").Inc()
				return
			}
			err := h.kube.delete(&User{Name: userToRemove})
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				metricRevokedHTTPTokensADMIN.WithLabelValues(username, "error").Inc()
				return
			} else {
				metricRevokedHTTPTokensADMIN.WithLabelValues(username, "success").Inc()
				c.String(http.StatusOK, "removed kube config for user (%s)", userToRemove)
				return
			}
		} else {
			c.String(http.StatusUnauthorized, "Ops.. username did not have permission to revoke other users)")
		}
	}
	// return default
	c.String(http.StatusInternalServerError, "Ops.. username did not have permission configured)")
}
