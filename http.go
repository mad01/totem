package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
	authorized.GET("/kube/config/create", h.handlerKubeConfig)
	authorized.GET("/kube/config/revoke", h.handlerKubeConfigRevoke) // todo: should be delete on same as create

	h.router.Run(fmt.Sprintf(":%d", h.config.Port))
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
				break
			}
			log().Infof(
				"generated kube config for cluster: (%s) cluster role: (%s) to (%s)",
				h.kube.cluster,
				user.ClusterRole,
				username,
			)
			c.String(http.StatusOK, cfg)
			break
		} else {
			c.String(http.StatusInternalServerError, "Ops.. username did not have access configured)")
			break
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
				break
			}
			err = h.kube.deleteServiceAccounts(username)
			if err != nil {
				c.String(
					http.StatusInternalServerError,
					"Ops.. failed to remove service account (%s) )", username,
				)
				break
			}
			c.String(http.StatusOK, "removed kube config for user (%s)", username)
			break
		} else {
			c.String(http.StatusInternalServerError, "Ops.. username did not have access configured)")
			break
		}
	}
}

//todo: change get kube config to get a new config every time.
//todo: change create kube config to add lables with username to service account and cluster role binding
//todo: change delete kube config to remove all service accounts and cluster role bindings matching username labels
