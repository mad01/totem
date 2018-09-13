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
	router := gin.Default()

	router.GET("/health", h.handlerHealth)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	authorized := router.Group("/api/", gin.BasicAuth(*h.config.GinAccounts))
	authorized.GET("/kube/config/create", h.handlerKubeConfig)
	authorized.GET("/kube/config/revoke", h.handlerKubeConfigRevoke) // todo: should be delete on same as create

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
	for _, user := range h.config.Users {
		if user.Name == username {
			cfg, err := h.kube.getServiceAccountKubeConfig(user.ClusterRole, username)
			if err != nil {
				c.String(http.StatusInternalServerError, err.Error())
				log().Error(err.Error())
				return
			}
			log().Infof(
				"generated kube config for cluster: (%s) cluster role: (%s) to (%s)",
				h.kube.cluster,
				user.ClusterRole,
				username,
			)
			c.String(http.StatusOK, cfg)
			return
		} else {
			c.String(http.StatusInternalServerError, "Ops.. username did not have access configured)")
			return
		}
	}
}

func (h *HttpServer) handlerKubeConfigRevoke(c *gin.Context) {
	username := c.MustGet(gin.AuthUserKey).(string)
	for _, user := range h.config.Users {
		if user.Name == username {
			err := h.kube.deleteClusterRoleBindings(username)
			if err != nil {
				c.String(
					http.StatusInternalServerError,
					"Ops.. failed to remove cluster role binding (%s) )", username,
				)
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
			c.String(http.StatusOK, "removed kube config for user (%s)", username)
			return
		} else {
			c.String(http.StatusInternalServerError, "Ops.. username did not have access configured)")
			return
		}
	}
}

//todo: change get kube config to get a new config every time.
//todo: change create kube config to add lables with username to service account and cluster role binding
//todo: change delete kube config to remove all service accounts and cluster role bindings matching username labels
