# totem
[![CircleCI](https://circleci.com/gh/mad01/totem.svg?style=svg)](https://circleci.com/gh/mad01/totem)

### Problem statement
1. Managment of kube configs when multiple orgs and teams are involved.
2. Not having access to configure and select a auth provider for the cluster. 
3. Having short lived kube configs
4. Having individual kube configs 
5. Having having the option to use different cluster roles for different individuals


### Solution
To allow the solution to run both when we have access to the master and can configure a auth provider and when not. Using the service accounts as a base for the indivdual kube configs and using the service account token and cert to generate a kube config. When the kube config have passed the allowed ttl the service account is removed and access is removed

### Usage

