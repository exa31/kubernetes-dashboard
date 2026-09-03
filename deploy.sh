#!/usr/bin/env bash
# KubeNexus 1-Click Kubernetes Deployment Automation Script (Bash)

set -e

ACTION="${1:-all}"
NAMESPACE="${2:-kubenexus}"
RELEASE_NAME="kubenexus"

function print_header() {
    echo -e "\n\033[1;36m========================================================\033[0m"
    echo -e "\033[1;36m  $1\033[0m"
    echo -e "\033[1;36m========================================================\033[0m\n"
}

function invoke_helm() {
    if command -v helm &> /dev/null; then
        helm "$@"
    else
        echo -e "\033[1;33mLocal helm CLI not detected. Using containerized Helm (alpine/helm)...\033[0m"
        docker run --rm \
            -v "$(pwd)/helm/kubenexus:/apps" \
            -v "$HOME/.kube:/root/.kube" \
            alpine/helm:3.17.0 "$@"
    fi
}

case "$ACTION" in
    build)
        print_header "BUILDING DOCKER IMAGES (MULTI-STAGE)"
        echo -e "\033[1;32m1/3 Building Backend (Go 1.25 Alpine)...\033[0m"
        docker build -t kubenexus-backend:latest ./be

        echo -e "\033[1;32m2/3 Building Frontend (Node 22 + Nginx)...\033[0m"
        docker build -t kubenexus-frontend:latest ./fe

        echo -e "\033[1;32m3/3 Building Database Migration Tool...\033[0m"
        docker build -t kubenexus-migrate:latest -f ./be/Dockerfile.migrate ./be
        ;;

    deploy)
        print_header "DEPLOYING TO KUBERNETES VIA HELM"
        kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
        invoke_helm upgrade --install "$RELEASE_NAME" ./helm/kubenexus --namespace "$NAMESPACE"
        ;;

    status)
        print_header "KUBENEXUS CLUSTER STATUS ($NAMESPACE)"
        kubectl get all,ingress,pvc,secrets -n "$NAMESPACE"
        ;;

    logs)
        print_header "BACKEND LOGS STREAM"
        kubectl logs -n "$NAMESPACE" -l app.kubernetes.io/component=backend -f --tail=100
        ;;

    all)
        print_header "STARTING FULL 1-CLICK KUBENEXUS DEPLOYMENT"
        echo -e "\033[1;33m[1/3] Building images...\033[0m"
        docker build -t kubenexus-backend:latest ./be
        docker build -t kubenexus-frontend:latest ./fe
        docker build -t kubenexus-migrate:latest -f ./be/Dockerfile.migrate ./be

        echo -e "\n\033[1;33m[2/3] Deploying via Helm...\033[0m"
        kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -
        invoke_helm upgrade --install "$RELEASE_NAME" ./helm/kubenexus --namespace "$NAMESPACE"

        echo -e "\n\033[1;33m[3/3] Checking cluster status...\033[0m"
        sleep 3
        kubectl get all,ingress -n "$NAMESPACE"
        ;;

    *)
        echo "Usage: $0 {build|deploy|status|logs|all} [namespace]"
        exit 1
        ;;
esac
