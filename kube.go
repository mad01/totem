package main

import (
	b64 "encoding/base64"
	"fmt"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/util/uuid"

	"github.com/mad01/totem/internal/try"

	"github.com/pkg/errors"
	"k8s.io/api/core/v1"
	rbac "k8s.io/api/rbac/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	annotation          = "k8s.io.totem/managed"
	annotationUsername  = "k8s.io.totem/username"
	annotationCreatedAt = "k8s.io.totem/created-at" // timeFormat
	timeFormat          = time.RFC3339
)

var (
	errorMissingCertData  = errors.New("missing cert data")
	errorMissingTokenData = errors.New("missing token data")
)

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

func (k *Kube) createClusterRoleBinding(clusterRole, username string, sa *v1.ServiceAccount) error {
	crb := &rbac.ClusterRoleBinding{
		ObjectMeta: meta_v1.ObjectMeta{
			Name: sa.Name,
			Annotations: map[string]string{
				annotation:          "",
				annotationCreatedAt: time.Now().Format(timeFormat),
			},
			Labels: map[string]string{
				annotationUsername: username,
			},
		},
		TypeMeta: meta_v1.TypeMeta{
			Kind:       "ClusterRole",
			APIVersion: "rbac.authorization.k8s.io/v1beta1",
		},
		RoleRef: rbac.RoleRef{
			APIGroup: rbac.GroupName,
			Kind:     "ClusterRole",
			Name:     clusterRole,
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

func (k *Kube) createServiceAccount(name, username string) (*v1.ServiceAccount, error) {
	sa := &v1.ServiceAccount{}
	sa.Name = name
	sa.Annotations = map[string]string{
		annotation:          "",
		annotationCreatedAt: time.Now().Format(timeFormat),
	}
	sa.Labels = map[string]string{
		annotationUsername: username,
	}

	return k.client.CoreV1().ServiceAccounts(k.serviceAccountNamespace).Create(sa)
}

func (k *Kube) deleteServiceAccount(username string) error {
	log().Infof("delete of service accounts matching label %s=%s", annotationUsername, username)
	return k.client.CoreV1().ServiceAccounts(k.serviceAccountNamespace).DeleteCollection(
		&meta_v1.DeleteOptions{},
		meta_v1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", annotationUsername, username),
		},
	)

}

func (k *Kube) deleteClusterRoleBinding(username string) error {
	log().Infof("delete of cluster role binding matching label %s=%s", annotationUsername, username)
	return k.client.RbacV1().ClusterRoleBindings().DeleteCollection(
		&meta_v1.DeleteOptions{},
		meta_v1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", annotationUsername, username),
		},
	)
}

func (k *Kube) getSecret(sa *v1.ServiceAccount) (*v1.Secret, error) {
	getFn := func(sa *v1.ServiceAccount) (*v1.Secret, error) {
		account, err := k.getServiceAccount(sa.Name)
		if err != nil {
			return nil, err
		}
		if len(account.Secrets) == 0 {
			return nil, errors.New("no secrets found in service account object")
		}
		secret, err := k.client.CoreV1().Secrets(k.serviceAccountNamespace).Get(
			account.Secrets[0].Name, meta_v1.GetOptions{},
		)
		if err != nil {
			return nil, err
		}

		if len(k.getSecretUserToken(secret)) < 10 {
			return nil, errorMissingTokenData
		}

		if len(k.getSecretCaCert(secret)) < 10 {
			return nil, errorMissingCertData
		}

		return secret, nil
	}

	var secret *v1.Secret
	err := try.Do(func(attempt int) (bool, error) {
		var err error
		secret, err = getFn(sa)
		if err != nil {
			time.Sleep(500 * time.Millisecond) // wait a bit
		}
		return attempt < 10, err
	})
	if err != nil {
		return nil, err
	}

	return secret, nil
}

func (k *Kube) getSecretCaCert(secret *v1.Secret) string {
	if certBytes, ok := secret.Data["ca.crt"]; ok {
		cert := b64.StdEncoding.EncodeToString(certBytes)
		return string(cert)
	}
	return ""
}

func (k *Kube) getSecretUserToken(secret *v1.Secret) string {
	if tokenB64, ok := secret.Data["token"]; ok {
		return string(tokenB64)
	}
	return ""
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

func (k *Kube) getServiceAccountKubeConfig(clusterRole, username string) (string, error) {
	name := fmt.Sprintf("%s", uuid.NewUUID())
	account, err := k.createServiceAccount(name, username)
	if errCheck(err) {
		return "", err
	}

	err = k.createClusterRoleBinding(clusterRole, username, account)
	if errCheck(err) {
		return "", err
	}

	secret, err := k.getSecret(account)
	if errCheck(err) {
		return "", err
	}

	cert := k.getSecretCaCert(secret)
	if len(cert) < 10 {
		return "", errorMissingCertData
	}

	token := k.getSecretUserToken(secret)
	if len(token) < 10 {
		return "", errorMissingTokenData
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
