#!/usr/bin/env bats
#

set -ef -o pipefail

function call-http() {
    result=$(http --check-status --ignore-stdin --timeout=2.5 -a "$1":qwerty123 "$2" http://localhost:8080/"$3" &>/dev/null)
    return $result
}

@test "get kube config alexander test" {
  call-http "alexander" "GET" "api/kubeconfig" "get kube config test"
  [ "$status" -eq 2 ]
}

@test "get kube config view test" {
  call-http "view" "GET" "api/kubeconfig" "get kube config test"
  [ "$status" -eq 2 ]
}

@test "get kube config edit test" {
  call-http "edit" "GET" "api/kubeconfig" "get kube config test"
  [ "$status" -eq 2 ]
}

@test "try get of kube config with unauthorized user" {
  http -a noaccess:qwerty123 --ignore-stdin --timeout=2.5 --check-status http://localhost:8080/api/kubeconfig &>/dev/null
  [ "$status" -eq 2 ]
}

@test "test to delete users config" {
  call-http "alexander" "DELETE" "api/kubeconfig" "delete kube config"
  [ "$status" -eq 2 ]
}

@test "test to delete other users config with admin user" {
  result=$(call-http "alexander" "GET" "metrics" "get metrics" | grep totem)
  if [[ $result == *"totem"* ]]; then
      exit 0
  else
      echo 1
  fi
}

@test "test revoke of other users configs" {
  result=$( call-http "alexander" "DELETE" "api/revoke/view" "test to revoke all configs for view with user alexander" )
  sleep 1
  after=$(kubectl get sa)

  if [[ $after == *"view"* ]]; then
      exit 1
  else
      exit 0
  fi
}
