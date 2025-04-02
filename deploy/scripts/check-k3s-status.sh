#!/bin/bash

# K3s集群状态检查脚本
# 用于检查集群内的Pod状态，以及验证节点和组件是否正常运行
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Function for printing progress messages
print_progress() {
    echo -e "${GREEN}[PROGRESS]${NC} $1"
}

# Function for printing warning messages
print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function for printing error messages
print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

if [ $# -lt 1 ]; then
    print_error "Usage: $0 <server_name>"
    exit 1
fi

# Read server configuration
print_progress "Reading server configuration..."
SERVER_CONFIG=$(cat deploy/config/servers.yaml)
SERVER_NAME=$1
SERVER_HOST=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .host" -)
SERVER_USER=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .user" -)
SERVER_REGION=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .region" -)

if [ -z "$SERVER_HOST" ]; then
    print_error "Server $SERVER_NAME not found in configuration"
    exit 1
fi

print_progress "Checking K3s status on server: $SERVER_NAME ($SERVER_HOST)"

# Check if k3s is installed
print_progress "Checking if k3s is installed..."
K3S_INSTALLED=$(ssh $SERVER_USER@$SERVER_HOST "command -v k3s > /dev/null && echo 'yes' || echo 'no'")

if [ "$K3S_INSTALLED" != "yes" ]; then
    print_error "k3s is not installed on the server"
    exit 1
fi

# Get k3s version
K3S_VERSION=$(ssh $SERVER_USER@$SERVER_HOST "k3s --version")
print_progress "K3s version: $K3S_VERSION"

# Check node status
print_progress "Checking node status..."
ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get nodes -o wide"

# Check pod status across all namespaces
print_progress "Checking pod status across all namespaces..."
ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get pods --all-namespaces"

# Check for non-running pods
print_progress "Checking for non-running pods..."
NOT_RUNNING_PODS=$(ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get pods --all-namespaces -o wide | grep -v 'Running' | grep -v 'Completed' | grep -v 'NAME'")

if [ -n "$NOT_RUNNING_PODS" ]; then
    print_warning "Some pods are not in Running state:"
    echo "$NOT_RUNNING_PODS"
    
    print_progress "Getting details for non-running pods..."
    # Extract namespace and pod name from the NOT_RUNNING_PODS output
    echo "$NOT_RUNNING_PODS" | while read -r line; do
        NAMESPACE=$(echo "$line" | awk '{print $1}')
        POD=$(echo "$line" | awk '{print $2}')
        
        if [ -n "$NAMESPACE" ] && [ -n "$POD" ]; then
            print_progress "Details for pod $NAMESPACE/$POD:"
            ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl describe pod $POD -n $NAMESPACE"
            print_progress "Logs for pod $NAMESPACE/$POD:"
            ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl logs $POD -n $NAMESPACE --tail=50 || echo 'No logs available'"
        fi
    done
else
    print_progress "All pods are in Running or Completed state! 👍"
fi

# Check services
print_progress "Checking services..."
ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get services --all-namespaces"

# Check ingress resources
print_progress "Checking ingress resources..."
ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get ingress --all-namespaces || echo 'No ingress resources found'"

# Check storage classes
print_progress "Checking storage classes..."
ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get storageclass"

# Get k3s service status
print_progress "Checking k3s service status..."
ssh $SERVER_USER@$SERVER_HOST "sudo systemctl status k3s | head -n 20"

# Check if ingress controller is deployed
print_progress "Checking if ingress controller is deployed..."
INGRESS_DEPLOYED=$(ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get pods -A | grep ingress")

if [ -n "$INGRESS_DEPLOYED" ]; then
    print_progress "Ingress controller is deployed:"
    echo "$INGRESS_DEPLOYED"
else
    print_warning "Ingress controller is not deployed"
    
    # Deploy ingress if in global region (should have better network connectivity)
    if [ "$SERVER_REGION" = "global" ]; then
        print_progress "Deploying ingress controller for global region..."
        ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.9.5/deploy/static/provider/cloud/deploy.yaml"
    fi
fi

# Overall assessment
print_progress "Overall K3s cluster assessment..."
NODE_READY=$(ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get nodes | grep Ready | wc -l")
TOTAL_PODS=$(ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get pods -A | grep -v NAME | wc -l")
RUNNING_PODS=$(ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get pods -A | grep Running | wc -l")
COMPLETED_PODS=$(ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get pods -A | grep Completed | wc -l")
NOT_READY_PODS=$((TOTAL_PODS - RUNNING_PODS - COMPLETED_PODS))

print_progress "Cluster Summary:"
echo "- Ready Nodes: $NODE_READY"
echo "- Total Pods: $TOTAL_PODS"
echo "- Running Pods: $RUNNING_PODS"
echo "- Completed Pods: $COMPLETED_PODS"
echo "- Non-Ready Pods: $NOT_READY_PODS"

if [ "$NODE_READY" -gt 0 ] && [ "$NOT_READY_PODS" -eq 0 ]; then
    print_progress "✅ Cluster appears to be healthy and ready!"
elif [ "$NODE_READY" -gt 0 ] && [ "$NOT_READY_PODS" -lt 3 ]; then
    print_warning "⚠️ Cluster is mostly operational with a few non-ready pods"
else
    print_error "❌ Cluster has issues that need to be addressed"
fi
