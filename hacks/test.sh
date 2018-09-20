#!/bin/bash -
set -o nounset # Treat unset variables as an error

function call-http() {
	if http --check-status --ignore-stdin --timeout=2.5 -a "$1":qwerty123 "$2" http://localhost:8080/"$3" &>/dev/null; then
		echo "OK! $4 with user $1"
	else
		case $? in
		2) echo 'Request timed out!' ;;
		3) echo 'Unexpected HTTP 3xx Redirection!' ;;
		4) echo 'HTTP 4xx Client Error!' ;;
		5) echo 'HTTP 5xx Server Error!' ;;
		6) echo 'Exceeded --max-redirects=<n> redirects!' ;;
		*) echo 'Other Error!' ;;
		esac
	fi
}

function get-test() {
	ARR=(
		"alexander"
		"view"
		"view"
		"edit"
	)
	for i in "${ARR[@]}"; do
		call-http "${i}" "GET" "api/kubeconfig" "get kube config test"
	done
	sleep 4
}

function get-test-unauthorized() {
	http -a noaccess:qwerty123 --ignore-stdin --timeout=2.5 --check-status http://localhost:8080/api/kubeconfig &>/dev/null
	if [[ $? == 4 ]]; then
		echo 'OK! unauthorized user did not have access'
	else
		echo 'Error! unauthorized user got access'
	fi
}

function delete-test() {
	ARR=(
		"alexander"
	)
	for i in "${ARR[@]}"; do
		call-http "${i}" "DELETE" "api/kubeconfig" "delete kube config test"
	done
}

function get-metrics() {
	call-http "alexander" "GET" "metrics" "get metrics test"
	result=$(http http://localhost:8080/metrics | grep totem)
	if [[ $result == *"totem"* ]]; then
		echo "OK! found totem prometheus metrics"
	else
		echo "Err! totem prometheus metrics not found"
	fi
}

function revoke-access() {
	call-http "alexander" "DELETE" "api/revoke/view" "test to revoke all configs for view with user alexander"
	sleep 1
	after=$(kubectl get sa)

	if [[ $after == *"view"* ]]; then
		echo "Err! totem prometheus metrics not found"
	else
		echo "OK! found totem prometheus metrics"
	fi
}

get-test
delete-test
revoke-access
get-test-unauthorized
get-metrics
