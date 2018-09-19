# totem
[![CircleCI](https://circleci.com/gh/mad01/totem.svg?style=svg)](https://circleci.com/gh/mad01/totem)
[![Docker Repository on Quay](https://quay.io/repository/mad01/totem/status "Docker Repository on Quay")](https://quay.io/repository/mad01/totem)

### Problem statement
1. Managment of kube configs when multiple orgs and teams are involved.
2. Not having access to configure and select a auth provider for the cluster. 
3. Having short lived kube configs
4. Having individual kube configs 
5. Having having the option to use different cluster roles for different individuals


### Solution
To allow the solution to run both when we have access to the master and can configure a auth provider and when not. Using the service accounts as a base for the indivdual kube configs and using the service account token and cert to generate a kube config. When the kube config have passed the allowed ttl the service account is removed and access is removed.

Creating a config. when you request a new config you get one generated with the configures lifetime. you can have multiple configs active at one time
Deleting a config. when you delete a config it will remove all configs created for your user. 
Adding a new user. you need to update the deployment config map with the new user, and recreate the pod


### Usage
get config for youre user
```bash
http -a username:pass GET http://example.com:8080/api/kubeconfig > config
KUBECONFIG=config kubectl get pods 
```


revoke config for your user
```bash
http -a username:pass DELETE http://example.com:8080/api/kubeconfig
```


#### config 
the config only contains the users currently. The way you can manage different
levels of access, is by defining a cluster role. The cluster role will then be 
bound to the service account for the user with a cluster role binding

example config were we use the defualt admin and view cluster roles. The 
alexander user will get bound to the admin cluster role, and the test user 
will be bound to the view cluster role
```yaml
---
users:
  - {name: alexander, clusterRole: admin, password: qwerty123}
  - {name: test, clusterRole: view, password: qwerty123}
```


#### deployment
start with the `template/deployment.yaml` as the base.
* `ADDR` fill in the address of the kubernetes api server
* `NAME` fill in the name of the cluster
* `VERSION` fill in the container version


#### running in/outside cluster
when running in the cluster: see `template/deployment.yaml` for example
* a service account
* a cluster role 
* a cluster role bindig for the service account

when running outside the cluster:
* a service account
* a cluster role 
* a cluster role bindig for the service account (the lazy way is to bind to the `cluster-admin` cluster role)
* a kube config generated from the service account secret to be able to access the cluster with the correct permissions (lazy way use the admin kube config)


#### cli controller flags
```
run the controller

Usage:
  totem controller [flags]

Flags:
  -a, --cluster.addr string       public dns to api cluster
  -c, --cluster.name string       name of k8s cluster (default "default")
  -u, --config string             path to config. config contains user/role mapping
  -h, --help                      help for controller
  -p, --http.port int             port to expose service on (default 8080)
  -i, --interval duration         the interval in which the cleanup of old token runs (default 1m0s)
  -k, --kube.config string        outside cluster path to kube config
  -n, --namespace string          ns where the service accounts and cluster role bindings is created (default "default")
  -l, --token.lifetime duration   the time that a kube config is valid for (default 1h0m0s)
  -v, --verbose                   verbose output
```
