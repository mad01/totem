#!/bin/bash
set -ef -o pipefail

# Check for dependencies
command -v jq >/dev/null 2>&1 || {
    echo >&2 "I require jq but it's not installed.  Aborting."
    exit 1
}
command -v kubectl >/dev/null 2>&1 || {
    echo >&2 "I require kubectl but it's not installed.  Aborting."
    exit 1
}
command -v mktemp >/dev/null 2>&1 || {
    echo >&2 "I require mktemp but it's not installed.  Aborting."
    exit 1
}

if [[ -z "$1" ]] || [[ -z "$2" ]] || [[ -z "$3" ]]; then
    echo "usage: $0 <service_account_name> <namespace> <rbac_access_level>"
    exit 1
fi

readonly SERVICE_ACCOUNT_NAME=$1
readonly NAMESPACE="$2"
readonly ACCESS_LEVEL="$3"
readonly TEMPDIR=$(mktemp -d)
readonly KUBECFG_FILE_NAME="$TEMPDIR/k8s-${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-conf"

readonly ADMIN_DOC=$(
    cat <<-EOM
Allows admin access, intended to be granted within a namespace using
a RoleBinding. If used in a RoleBinding, allows read/write access to
most resources in a namespace, including the ability to create roles
and rolebindings within the namespace. It does not allow write
access to resource quota or to the namespace itself.
EOM
)

readonly EDIT_DOC=$(
    cat <<EOM
Allows read/write access to most objects in a namespace. It does not
allow viewing or modifying roles or rolebindings.
EOM
)

readonly VIEW_DOC=$(
    cat <<EOM
Allows read-only access to see most objects in a namespace. It does
not allow viewing roles or rolebindings. It does not allow viewing
secrets, since those are escalating.
EOM
)

check_access_level() {
    case "${ACCESS_LEVEL}" in
    "view")
        echo "adding access level (${ACCESS_LEVEL})"
        echo "$VIEW_DOC"
        ;;

    "edit")
        echo "adding access level (${ACCESS_LEVEL})"
        echo "$EDIT_DOC"
        ;;

    "admin")
        echo "adding access level (${ACCESS_LEVEL})"
        echo "$ADMIN_DOC"
        ;;

    *)
        echo ""
        echo "Ops.. level not supported: (${ACCESS_LEVEL})"
        echo ""
        echo "supported levels are <view|edit|admin>"
        echo ""
        echo "view: $VIEW_DOC"
        echo ""
        echo "edit: $EDIT_DOC"
        echo ""
        echo "admin: $ADMIN_DOC"
        echo ""
        exit 1
        ;;
    esac
}

create_service_account() {
    echo -e "\\nCreating a service account: ${SERVICE_ACCOUNT_NAME} on namespace: ${NAMESPACE}"
    kubectl create sa "${SERVICE_ACCOUNT_NAME}" --namespace "${NAMESPACE}"
}

get_secret_name_from_service_account() {
    echo -e "\\nGetting secret of service account ${SERVICE_ACCOUNT_NAME}-${NAMESPACE}"
    SECRET_NAME=$(kubectl get sa "${SERVICE_ACCOUNT_NAME}" --namespace "${NAMESPACE}" -o json | jq -r '.secrets[].name')
    echo "Secret name: ${SECRET_NAME}"
}

extract_ca_crt_from_secret() {
    echo -e -n "\\nExtracting ca.crt from secret..."
    kubectl get secret "${SECRET_NAME}" --namespace "${NAMESPACE}" -o json | jq \
        -r '.data["ca.crt"]' | base64 -d >"${TEMPDIR}/ca.crt"
    printf "done"
}

get_user_token_from_secret() {
    echo -e -n "\\nGetting user token from secret..."
    USER_TOKEN=$(kubectl get secret "${SECRET_NAME}" \
        --namespace "${NAMESPACE}" -o json | jq -r '.data["token"]' | base64 -d)
    printf "done"
}

set_kube_config_values() {
    context=$(kubectl config current-context)
    echo -e "\\nSetting current context to: $context"

    CLUSTER_NAME=$(kubectl config get-contexts "$context" | awk '{print $3}' | tail -n 1)
    echo "Cluster name: ${CLUSTER_NAME}"

    ENDPOINT=$(kubectl config view \
        -o jsonpath="{.clusters[?(@.name == \"${CLUSTER_NAME}\")].cluster.server}")
    echo "Endpoint: ${ENDPOINT}"

    # Set up the config
    echo -e "\\nPreparing k8s-${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-conf"
    echo -n "Setting a cluster entry in kubeconfig..."
    kubectl config set-cluster "${CLUSTER_NAME}" \
        --kubeconfig="${KUBECFG_FILE_NAME}" \
        --server="${ENDPOINT}" \
        --certificate-authority="${TEMPDIR}/ca.crt" \
        --embed-certs=true

    echo -n "Setting token credentials entry in kubeconfig..."
    kubectl config set-credentials \
        "${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
        --kubeconfig="${KUBECFG_FILE_NAME}" \
        --token="${USER_TOKEN}"

    echo -n "Setting a context entry in kubeconfig..."
    kubectl config set-context \
        "${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
        --kubeconfig="${KUBECFG_FILE_NAME}" \
        --cluster="${CLUSTER_NAME}" \
        --user="${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
        --namespace="${NAMESPACE}"

    echo -n "Setting the current-context in the kubeconfig file..."
    kubectl config use-context "${SERVICE_ACCOUNT_NAME}-${NAMESPACE}-${CLUSTER_NAME}" \
        --kubeconfig="${KUBECFG_FILE_NAME}"
    printf "done"
}

set_kube_clusterrolebinding() {
    kubectl create clusterrolebinding \
        --serviceaccount="$NAMESPACE:$SERVICE_ACCOUNT_NAME" \
        --clusterrole="$ACCESS_LEVEL" "sa-$SERVICE_ACCOUNT_NAME-$ACCESS_LEVEL"
}

create_service_account
get_secret_name_from_service_account
extract_ca_crt_from_secret
get_user_token_from_secret
set_kube_config_values
check_access_level
set_kube_clusterrolebinding

echo -e "\\nAll done! Test with:"
echo "KUBECONFIG=${KUBECFG_FILE_NAME} kubectl get pods"
