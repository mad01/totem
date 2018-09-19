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
    echo "---------------------"
    ARR=(
        "alexander"
        "alexander"
        "view"
        "view"
        "view"
        "edit"
        "noaccess"
    )
    for i in "${ARR[@]}"; do
        call-http "${i}" "GET" "api/kubeconfig" "get kube config test"
        sleep 1
    done
    echo "---------------------"
}

function delete-test() {
    echo "---------------------"
    ARR=(
        "alexander"
    )
    for i in "${ARR[@]}"; do
        call-http "${i}" "DELETE" "api/kubeconfig" "delete kube config test"
        sleep 1
    done
    echo "---------------------"
}

function get-metrics() {
    echo "---------------------"
    call-http "alexander" "GET" "metrics" "get metrics test"
    echo ""
    http http://localhost:8080/metrics | grep totem
    echo "---------------------"
}

function revoke-access() {
    echo "---------------------"
    echo "test revoke access for user view"
    kubectl get sa
    echo ""
    call-http "admin" "DELETE" "api/revoke/view" "test revoke of access"
    echo ""
    sleep 1
    kubectl get sa
    echo "---------------------"
}

get-test # test get of config
echo ""
sleep 4
delete-test # test delete of users configs
echo ""
revoke-access # test revoke access for user
echo ""
get-metrics # get totem metrics