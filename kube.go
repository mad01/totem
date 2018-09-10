package main

import (
	b64 "encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const annotation = "k8s.io.totem/managed"
const annotationCreatedAt = "k8s.io.totem/created-at" // timeFormat
const timeFormat = time.RFC3339

type kubecfg struct {
	cert        string
	serverUrl   string
	clusterName string
	user        string
	token       string
}

type Kube struct {
	client                  *kubernetes.Clientset
	restConfig              *rest.Config
	serviceAccountNamespace string
	cluster                 string
}

func (k *Kube) createClusterRoleBinding(accessLevel string, sa *v1.ServiceAccount) error {
	// access level should be for a start view/edit/admin
	// options for the features is to add a config file with the allowed options
	// and then it's expected that we are bootstrapping the cluster with the needed
	// cluster role bindings to

	crb := &rbac.ClusterRoleBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: fmt.Sprintf("%s-%s", sa.Name, accessLevel),
			Annotations: map[string]string{
				annotation:          "",
				annotationCreatedAt: time.Now().Format(timeFormat),
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

func (k *Kube) createServiceAccount(name string) (*v1.ServiceAccount, error) {
	sa := &v1.ServiceAccount{}
	sa.Name = name
	sa.Annotations = map[string]string{
		annotation:          "",
		annotationCreatedAt: time.Now().Format(timeFormat),
	}

	return k.client.CoreV1().ServiceAccounts(k.serviceAccountNamespace).Create(sa)
}

func (k *Kube) deleteServiceAccount(name string) error {
	return k.client.CoreV1().ServiceAccounts(k.serviceAccountNamespace).Delete(name, &meta_v1.DeleteOptions{})
}

func (k *Kube) deleteClusterRoleBinding(name string) error {
	return k.client.RbacV1().ClusterRoleBindings().Delete(name, &meta_v1.DeleteOptions{})
}

func (k *Kube) getSecret(sa *v1.ServiceAccount) (*v1.Secret, error) {
	account, err := k.getServiceAccount(sa.Name)
	if err != nil {
		return nil, err
	}
	if len(account.Secrets) == 0 {
		return nil, errors.New("no secrets found in service account object")
	}
	return k.client.CoreV1().Secrets(k.serviceAccountNamespace).Get(
		account.Secrets[0].Name, meta_v1.GetOptions{},
	)

}

func (k *Kube) getSecretCaCert(secret *v1.Secret) string {
	if certBytes, ok := secret.Data["ca.crt"]; ok {
		cert := b64.StdEncoding.EncodeToString(certBytes)
		return string(cert)
	}
	return ""
}

func (k *Kube) getSecretUserToken(secret *v1.Secret) (string, error) {
	if tokenB64, ok := secret.Data["token"]; ok {
		return string(tokenB64), nil
	}
	return "", nil
}

func (k *Kube) getClusterRoleBindingList() (*rbac.ClusterRoleBindingList, error) {
	return k.client.RbacV1().ClusterRoleBindings().List(meta_v1.ListOptions{})
}

func (k *Kube) getServiceAccountList() (*v1.ServiceAccountList, error) {
	return k.client.CoreV1().ServiceAccounts(k.serviceAccountNamespace).List(meta_v1.ListOptions{})
}

func (k *Kube) getServiceAccount(name string) (*v1.ServiceAccount, error) {
	return k.client.CoreV1().ServiceAccounts(k.serviceAccountNamespace).Get(name, meta_v1.GetOptions{})
}

func (k *Kube) getServiceAccountKubeConfig(accessLevel, name string) (string, error) {
	account, err := k.createServiceAccount(name)
	if errCheck(err) {
		return "", err
	}

	err = k.createClusterRoleBinding(accessLevel, account)
	if errCheck(err) {
		return "", err
	}

	time.Sleep(time.Second * 2) //todo: add some retry login for getting secret to not have to sleep
	secret, err := k.getSecret(account)
	if errCheck(err) {
		return "", err
	}

	cert := k.getSecretCaCert(secret)
	token, err := k.getSecretUserToken(secret)
	if errCheck(err) {
		return "", err
	}

	cfg := &kubecfg{}
	cfg.user = account.Name
	cfg.token = token
	cfg.cert = cert
	cfg.clusterName = k.cluster
	cfg.serverUrl = k.restConfig.Host

	return k.generateKubeConfig(cfg), nil
}

func (k *Kube) generateKubeConfig(cfg *kubecfg) string {
	template := `
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
		"{cert-data}", cfg.cert,
		"{server-url}", cfg.serverUrl,
		"{clusterName}", cfg.clusterName,
		"{user}", cfg.user,
		"{token}", cfg.token,
		"{namespace}", k.serviceAccountNamespace,
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
		client:     client,
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

func errCheck(err error) bool {
	if err != nil {
		return true
	}
	return false
}
