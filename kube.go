package main

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Kube struct {
	client *kubernetes.Clientset
}

func newKube(kubeconfig string) *Kube {
	client, err := K8sGetClient(kubeconfig)
	if err != nil {
		log().Fatalf(err.Error())
	}

	k := &Kube{
		client: client,
	}
	return k
}

func K8sGetClientConfig(kubeconfig string) (*rest.Config, error) {
	if kubeconfig != "" {
		return clientcmd.BuildConfigFromFlags("", kubeconfig)
	}
	return rest.InClusterConfig()
}

func K8sGetClient(kubeconfig string) (*kubernetes.Clientset, error) {
	config, err := K8sGetClientConfig(kubeconfig)
	if err != nil {
		return nil, err
	}

	// Construct the Kubernetes client
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return client, nil
}
