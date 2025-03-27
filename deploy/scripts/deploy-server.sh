#!/bin/bash

# Server deployment script
set -e

# Read server configuration
SERVER_CONFIG=$(cat deploy/config/servers.yaml)

# Parse server information
SERVER_NAME=$1
SERVER_HOST=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .host" -)
SERVER_USER=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .user" -)
SERVER_K3S_VERSION=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .k3s_version" -)
SERVER_REGION=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .region" -)

if [ -z "$SERVER_HOST" ]; then
    echo "Error: Server $SERVER_NAME not found in configuration"
    exit 1
fi

echo "Deploying to server: $SERVER_NAME ($SERVER_HOST), user: $SERVER_USER, k3s version: $SERVER_K3S_VERSION, region: $SERVER_REGION"

# 1. Install necessary dependencies
ssh $SERVER_USER@$SERVER_HOST "sudo apt-get update && sudo apt-get install -y curl wget"

# 2. Install k3s
# Use Chinese mirror for servers in China region
# curl -sfL https://rancher-mirror.rancher.cn/k3s/k3s-install.sh | INSTALL_K3S_MIRROR=cn sh -
if [ "$SERVER_REGION" = "cn" ]; then
    echo "Using China mirror for k3s installation"
    ssh $SERVER_USER@$SERVER_HOST " \
        curl -sfL https://rancher-mirror.rancher.cn/k3s/k3s-install.sh | \
        INSTALL_K3S_MIRROR=cn \
        INSTALL_K3S_VERSION=$SERVER_K3S_VERSION \
        K3S_TOKEN=$(uuidgen) \
        sudo sh -
    "
else
    echo "Using global mirror for k3s installation"
    ssh $SERVER_USER@$SERVER_HOST " \
        curl -sfL https://get.k3s.io | \
        INSTALL_K3S_VERSION=$SERVER_K3S_VERSION \
        K3S_TOKEN=$(uuidgen) \
        sudo sh -
    "
fi

# 3. Configure k3s
ssh $SERVER_USER@$SERVER_HOST " \
    sudo mkdir -p /etc/rancher/k3s \
    && echo 'write-kubeconfig-mode=\"644\"' | sudo tee /etc/rancher/k3s/config.yaml \
    && echo 'cluster-cidr=10.42.0.0/16' | sudo tee -a /etc/rancher/k3s/config.yaml
"

# 4. Configure Docker registry (using actual registry address)
REGISTRY_URL=$(grep "REGISTRY" Makefile | head -n1 | cut -d'=' -f2 | tr -d ' ')
if [ -z "$REGISTRY_URL" ]; then
    REGISTRY_URL="docker.io"  # Default to DockerHub
fi

ssh $SERVER_USER@$SERVER_HOST " \
    sudo mkdir -p /etc/rancher/k3s/registries.d \
    && echo '{\"mirrors\": {\"$REGISTRY_URL\": {\"endpoint\": [\"https://$REGISTRY_URL\"]}}}' | sudo tee /etc/rancher/k3s/registries.d/registry.conf
"

# 5. Configure Ingress
ssh $SERVER_USER@$SERVER_HOST " \
    sudo k3s kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.9.5/deploy/static/provider/cloud/deploy.yaml
"

# 6. Get kubeconfig
ssh $SERVER_USER@$SERVER_HOST " \
    sudo cp /etc/rancher/k3s/k3s.yaml /home/$SERVER_USER/k3s.yaml \
    && sudo chown $SERVER_USER:$SERVER_USER /home/$SERVER_USER/k3s.yaml \
    && sudo sed -i 's/127.0.0.1/$SERVER_HOST/g' /home/$SERVER_USER/k3s.yaml
"

# 7. Configure local kubectl
mkdir -p ~/.kube
scp $SERVER_USER@$SERVER_HOST:/home/$SERVER_USER/k3s.yaml ~/.kube/config-$SERVER_NAME
echo "Kubeconfig saved to ~/.kube/config-$SERVER_NAME"
echo "To use this config, run: export KUBECONFIG=~/.kube/config-$SERVER_NAME"

# 8. Verify installation
ssh $SERVER_USER@$SERVER_HOST " \
    sudo k3s kubectl get nodes \
    && sudo k3s kubectl get pods --all-namespaces
"

echo "Server $SERVER_NAME deployment complete!"
