// Package k8s containers adapters to watch k8s api servers.
package k8s

import (
	"time"

	"github.com/mad01/totem/internal/workgroup"
	"github.com/sirupsen/logrus"

	"k8s.io/api/core/v1"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// WatchServices creates a SharedInformer for v1.Services and registers it with g.
func WatchServices(g *workgroup.Group, client *kubernetes.Clientset, log logrus.FieldLogger, rs ...cache.ResourceEventHandler) {
	watch(g, client.CoreV1().RESTClient(), log, "services", new(v1.Service), rs...)
}

// WatchEndpoints creates a SharedInformer for v1.Endpoints and registers it with g.
func WatchEndpoints(g *workgroup.Group, client *kubernetes.Clientset, log logrus.FieldLogger, rs ...cache.ResourceEventHandler) {
	watch(g, client.CoreV1().RESTClient(), log, "endpoints", new(v1.Endpoints), rs...)
}

// WatchIngress creates a SharedInformer for v1beta1.Ingress and registers it with g.
func WatchIngress(g *workgroup.Group, client *kubernetes.Clientset, log logrus.FieldLogger, rs ...cache.ResourceEventHandler) {
	watch(g, client.ExtensionsV1beta1().RESTClient(), log, "ingresses", new(v1beta1.Ingress), rs...)
}

func watch(g *workgroup.Group, c cache.Getter, log logrus.FieldLogger, resource string, objType runtime.Object, rs ...cache.ResourceEventHandler) {
	lw := cache.NewListWatchFromClient(c, resource, v1.NamespaceAll, fields.Everything())
	sw := cache.NewSharedInformer(lw, objType, 30*time.Minute)
	for _, r := range rs {
		sw.AddEventHandler(r)
	}
	g.Add(func(stop <-chan struct{}) {
		log := log.WithField("resource", resource)
		log.Println("started")
		defer log.Println("stopped")
		sw.Run(stop)
	})
}
