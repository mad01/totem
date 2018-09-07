package main

import (
	"fmt"
	"k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	rbac "k8s.io/api/rbac/v1"
	"strings"
	"time"
)

const annotation = "k8s.io.totem/managed"
const annotationCreadtedAt = "k8s.io.totem/created-at" // should be a timestamp of

type kubecfg struct {
	certData string
	serverUrl string
	clusterName string
	user string
	token string
}


type Kube struct {
	client *kubernetes.Clientset
	restConfig *rest.Config
	serviceAccountNamespace string
}

func (k *Kube) createClusterRoleBinding(accessLevel string, sa *v1.ServiceAccount) error {
	// access level should be for a start view/edit/admin
	// options for the features is to add a config file with the allowed options
	// and then it's expected that we are bootstrapping the cluster with the needed
	// cluster role bindings to

	crb := &rbac.ClusterRoleBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", accessLevel, sa.Name),
			Annotations: map[string]string{
				annotation: "",
				annotationCreadtedAt: time.Now().String(),
			},
		},
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "ClusterRoleBinding",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     accessLevel,
		},
		Subjects: []rbac.Subject{
			{
				Kind:      rbac.ServiceAccountKind,
				Namespace: k.serviceAccountNamespace,
				Name:      sa.Name,
			},
		},
	}

	k.client.RbacV1().ClusterRoleBindings().Create(crb)
	return nil
}

func (k *Kube) createServiceAccount() (*v1.ServiceAccount, error){
	sa := &v1.ServiceAccount{}
	sa.Name = fmt.Sprintf("%s", uuid.NewUUID())
	sa.Annotations = map[string]string{
		annotation: "",
		annotationCreadtedAt: time.Now().String(),
	}

	return k.client.CoreV1().ServiceAccounts(k.serviceAccountNamespace).Create(sa)
}

func (k *Kube) getSecret(namespace string, name string) (*v1.Secret, error) {
	// todo: implement
	return nil, nil
}

func (k *Kube) getSecretCaCert(secret *v1.Secret) string {
	// todo: implement
	return ""
}

func (k *Kube) getSecretUserToken(secret *v1.Secret) string {
	// todo: implement
	return ""
}

func (k *Kube) getServiceAccountList(namespace string) (*v1.ServiceAccountList, error) {
	return k.client.CoreV1().ServiceAccounts(namespace).List(meta_v1.ListOptions{})
}

func (k *Kube) getServiceAccount(namespace string, name string) (*v1.ServiceAccount, error) {
	return k.client.CoreV1().ServiceAccounts(namespace).Get(name, meta_v1.GetOptions{})
}


func (k *Kube) getServiceAccountKubeConfig() (string, error) {
	// to all the calls here to create everything to be able to return a kubeconfig
	// todo: implement
	return "", nil
}

func (k *Kube) generateKubeConfig(cfg *kubecfg) string {
template :=`
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: {cert-data}
    server: {server-url} 
  name: {clusterName}
contexts:
- context:
    cluster: {clusterName} 
    namespace: {namespace}
    user: {user}
  name: {user}
current-context: {user}
kind: Config
preferences: {}
users:
- name: {user}
  user:
    token: {token}
`

	var replacer = strings.NewReplacer(
		"{cert-data}", cfg.certData,
		"{server-url}", cfg.serverUrl,
		"{clusterName}", cfg.clusterName,
		"{user}", cfg.user,
		"{token}", cfg.token,
	)
	str := replacer.Replace(template)
	return str
}



func newKube(kubeconfig string) *Kube {
	client, err := K8sGetClient(kubeconfig)
	if err != nil {
		log().Fatalf(err.Error())
	}

	restClient, err := K8sGetClientConfig(kubeconfig)
	if err != nil {
		log().Fatalf(err.Error())
	}

	k := &Kube{
		client: client,
		restConfig: restClient,
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
