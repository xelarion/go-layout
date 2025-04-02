#!/bin/bash

# K3s Deployment Script
# This script handles the deployment of K3s to a server with optimizations for different regions
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

# Function to pull an image with fallbacks
pull_with_fallback() {
    image_name=$1
    image_tag=$2
    
    # List of mirror registries to try, in order of priority
    registries=(
        "registry.cn-hangzhou.aliyuncs.com"
        "registry.cn-beijing.aliyuncs.com"
        "registry.cn-shanghai.aliyuncs.com"
        "registry.cn-shenzhen.aliyuncs.com"
        "docker.m.daocloud.io"
        "dockerhub.mirrors.nwafu.edu.cn"
    )
    
    # Mapping path for special images
    if [[ $image_name == rancher/* ]]; then
        path="rancher/$(echo $image_name | cut -d'/' -f2)"
    elif [[ $image_name == registry.k8s.io/* ]]; then
        path="google_containers/$(echo $image_name | cut -d'/' -f2-)"
    else
        path=$image_name
    fi
    
    # Try to pull the image from mirrors
    for registry in "${registries[@]}"; do
        print_progress "Trying to pull from $registry: $path:$image_tag"
        if docker pull "$registry/$path:$image_tag" 2>/dev/null; then
            print_progress "Successfully pulled from $registry"
            docker tag "$registry/$path:$image_tag" "$image_name:$image_tag"
            return 0
        else
            print_warning "Failed to pull from $registry, trying next mirror"
        fi
    done
    
    # If all mirrors fail, try pulling directly from original source
    print_warning "All mirrors failed, trying to pull directly: $image_name:$image_tag"
    if docker pull "$image_name:$image_tag" 2>/dev/null; then
        print_progress "Successfully pulled image directly: $image_name:$image_tag"
        return 0
    fi
    
    # Special handling for alternative versions of certain images
    if [[ $image_name == "rancher/klipper-helm" && $image_tag == "v0.9.4" ]]; then
        alt_tags=("v0.9.4-build20250113" "v0.9.3" "v0.8.0")
        for alt_tag in "${alt_tags[@]}"; do
            print_warning "Trying alternative version: $image_name:$alt_tag"
            for registry in "${registries[@]}"; do
                if docker pull "$registry/$path:$alt_tag" 2>/dev/null; then
                    print_progress "Successfully pulled alternative: $registry/$path:$alt_tag"
                    docker tag "$registry/$path:$alt_tag" "$image_name:$image_tag"
                    return 0
                fi
            done
            
            if docker pull "$image_name:$alt_tag" 2>/dev/null; then
                print_progress "Successfully pulled alternative directly: $image_name:$alt_tag"
                docker tag "$image_name:$alt_tag" "$image_name:$image_tag"
                return 0
            fi
        done
    fi
    
    print_error "Failed to pull image $image_name:$image_tag and its alternatives"
    return 1
}

# Display usage information
show_usage() {
    echo "Usage: $0 <command> <server_name> [options]"
    echo ""
    echo "Commands:"
    echo "  install         Install K3s on the server"
    echo "  uninstall       Uninstall K3s from the server"
    echo "  restart         Restart K3s service"
    echo "  status          Check K3s status"
    echo "  preload-images  Preload K3s images to the server"
    echo ""
    echo "Options:"
    echo "  --version       K3s version to install (default: v1.31.6+k3s1)"
    echo "  --no-preload    Skip image preloading during installation"
    echo ""
    echo "Examples:"
    echo "  $0 install muggleday"
    echo "  $0 install muggleday --version v1.31.6+k3s1"
    echo "  $0 status muggleday"
    exit 1
}

# Check if at least two arguments are provided
if [ $# -lt 2 ]; then
    show_usage
fi

COMMAND=$1
SERVER_NAME=$2
shift 2

# Parse additional options
K3S_VERSION="v1.31.6+k3s1"
PRELOAD_IMAGES=true

while [[ $# -gt 0 ]]; do
    case $1 in
        --version)
            K3S_VERSION="$2"
            shift 2
            ;;
        --no-preload)
            PRELOAD_IMAGES=false
            shift
            ;;
        *)
            print_error "Unknown option: $1"
            show_usage
            ;;
    esac
done

# Read server configuration
print_progress "Reading server configuration..."
SERVER_CONFIG=$(cat deploy/config/servers.yaml)
SERVER_HOST=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .host" -)
SERVER_USER=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .user" -)
SERVER_REGION=$(echo "$SERVER_CONFIG" | yq e ".servers[] | select(.name == \"$SERVER_NAME\") | .region" -)

if [ -z "$SERVER_HOST" ]; then
    print_error "Server $SERVER_NAME not found in configuration"
    exit 1
fi

print_progress "Target server: $SERVER_NAME ($SERVER_HOST)"

# Function to preload K3s images
preload_images() {
    print_progress "Preloading K3s images for server: $SERVER_NAME"
    
    # Create a temporary directory for images
    TEMP_DIR=$(mktemp -d)
    mkdir -p $TEMP_DIR/images
    cd $TEMP_DIR
    
    # Essential K3s images list
    print_progress "Preparing essential K3s images..."
    cat > essential-images.txt << EOF
rancher/mirrored-pause:3.6
rancher/klipper-helm:v0.9.4
rancher/mirrored-metrics-server:v0.6.3
EOF
    
    # Download all images in the list
    print_progress "Downloading images locally..."
    mkdir -p images
    
    success_count=0
    failed_count=0
    total_count=$(wc -l < essential-images.txt)
    
    if [ "$SERVER_REGION" = "cn" ]; then
        print_progress "Using China mirror for downloading images..."
        
        # Use the improved pulling function
        while read image; do
            image_name=$(echo $image | cut -d':' -f1)
            image_tag=$(echo $image | cut -d':' -f2)
            
            if pull_with_fallback "$image_name" "$image_tag"; then
                # Export as tar package
                image_filename=$(echo $image | tr '/:' '._')
                print_progress "Saving: $image to $image_filename.tar"
                docker save $image -o images/$image_filename.tar
                success_count=$((success_count + 1))
            else
                failed_count=$((failed_count + 1))
                print_error "Failed to pull image: $image"
            fi
        done < essential-images.txt
    else
        # For global regions, pull original images directly
        while read image; do
            print_progress "Pulling: $image"
            if docker pull $image; then
                # Export as tar package
                image_filename=$(echo $image | tr '/:' '._')
                print_progress "Saving: $image to $image_filename.tar"
                docker save $image -o images/$image_filename.tar
                success_count=$((success_count + 1))
            else
                failed_count=$((failed_count + 1))
                print_error "Failed to pull image: $image"
            fi
        done < essential-images.txt
    fi
    
    print_progress "Download statistics: Success $success_count / Total $total_count (Failed: $failed_count)"
    
    if [ $success_count -eq 0 ]; then
        print_error "No images were successfully downloaded, exiting"
        exit 1
    fi
    
    if [ $failed_count -gt 0 ]; then
        print_warning "$failed_count images failed to download, but will continue with the successfully downloaded ones"
    fi
    
    # Create archive
    print_progress "Creating archive of all images..."
    tar czf k3s-images.tar.gz images/
    
    # Upload to server
    print_progress "Uploading images to $SERVER_NAME..."
    scp k3s-images.tar.gz $SERVER_USER@$SERVER_HOST:/tmp/
    
    # Extract and load images on the server
    print_progress "Loading images on $SERVER_NAME..."
    ssh $SERVER_USER@$SERVER_HOST "mkdir -p ~/k3s-images && tar xzf /tmp/k3s-images.tar.gz -C ~/k3s-images && cd ~/k3s-images && for img in images/*.tar; do echo 'Loading \$img...'; sudo ctr image import \$img || echo 'Failed to import \$img'; done"
    
    # Clean up local temporary files
    print_progress "Cleaning up local temporary files..."
    cd - > /dev/null
    rm -rf $TEMP_DIR
    
    print_progress "Image preloading completed"
}

# Function to check K3s status
check_status() {
    print_progress "Checking K3s status on server: $SERVER_NAME"
    
    # Check if k3s is installed
    print_progress "Checking if k3s is installed..."
    K3S_INSTALLED=$(ssh $SERVER_USER@$SERVER_HOST "command -v k3s > /dev/null && echo 'yes' || echo 'no'")
    
    if [ "$K3S_INSTALLED" != "yes" ]; then
        print_error "k3s is not installed on the server"
        exit 1
    fi
    
    # Get k3s version
    print_progress "K3s version:"
    ssh $SERVER_USER@$SERVER_HOST "sudo k3s --version"
    
    # Check node status
    print_progress "Checking node status..."
    ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get nodes -o wide"
    
    # Check pod status
    print_progress "Checking pod status across all namespaces..."
    ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get pods --all-namespaces"
    
    # Check for non-running pods
    print_progress "Checking for non-running pods..."
    NON_RUNNING_PODS=$(ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get pods --all-namespaces | grep -v Running | grep -v Completed | grep -v NAME")
    
    if [ -n "$NON_RUNNING_PODS" ]; then
        print_warning "Some pods are not in Running state:"
        echo "$NON_RUNNING_PODS"
    else
        print_progress "All pods are running or completed successfully"
    fi
    
    # Check services
    print_progress "Checking services..."
    ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get services --all-namespaces"
    
    # Check ingress resources
    print_progress "Checking ingress resources..."
    ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get ingress --all-namespaces 2>/dev/null || echo 'No resources found'"
    
    # Check storage classes
    print_progress "Checking storage classes..."
    ssh $SERVER_USER@$SERVER_HOST "sudo k3s kubectl get storageclasses"
    
    # Check k3s service status
    print_progress "Checking k3s service status..."
    ssh $SERVER_USER@$SERVER_HOST "sudo systemctl status k3s | head -n 20"
    
    print_progress "Status check completed"
}

# Function to install K3s
install_k3s() {
    print_progress "Installing K3s version $K3S_VERSION on server: $SERVER_NAME"
    
    # Check if k3s is already installed
    K3S_INSTALLED=$(ssh $SERVER_USER@$SERVER_HOST "command -v k3s > /dev/null && echo 'yes' || echo 'no'")
    
    if [ "$K3S_INSTALLED" = "yes" ]; then
        print_warning "K3s is already installed on the server"
        CURRENT_VERSION=$(ssh $SERVER_USER@$SERVER_HOST "sudo k3s --version | head -n 1")
        print_warning "Current version: $CURRENT_VERSION"
        
        read -p "Do you want to continue and reinstall K3s? (y/n): " CONTINUE
        if [ "$CONTINUE" != "y" ]; then
            print_progress "Installation aborted"
            exit 0
        fi
        
        print_progress "Uninstalling existing K3s installation..."
        ssh $SERVER_USER@$SERVER_HOST "sudo /usr/local/bin/k3s-uninstall.sh"
    fi
    
    # Preload images if enabled
    if [ "$PRELOAD_IMAGES" = true ]; then
        preload_images
    else
        print_progress "Skipping image preloading as requested"
    fi
    
    # Generate a unique token for K3s
    K3S_TOKEN=$(openssl rand -hex 16)
    print_progress "Generated K3s token"
    
    # Prepare registry mirrors configuration for China region
    if [ "$SERVER_REGION" = "cn" ]; then
        print_progress "Preparing registry mirrors configuration for China region..."
        
        # Create registry configuration
        cat > registry-mirrors.yaml << EOF
mirrors:
  docker.io:
    endpoint:
      - "https://registry.cn-hangzhou.aliyuncs.com"
      - "https://docker.m.daocloud.io"
  registry.k8s.io:
    endpoint:
      - "https://registry.cn-hangzhou.aliyuncs.com/google_containers"
  quay.io:
    endpoint:
      - "https://quay.mirrors.ustc.edu.cn"
  gcr.io:
    endpoint:
      - "https://registry.cn-hangzhou.aliyuncs.com/google_containers"
EOF
        
        # Upload registry configuration
        print_progress "Uploading registry configuration..."
        ssh $SERVER_USER@$SERVER_HOST "sudo mkdir -p /etc/rancher/k3s"
        scp registry-mirrors.yaml $SERVER_USER@$SERVER_HOST:/tmp/
        ssh $SERVER_USER@$SERVER_HOST "sudo mv /tmp/registry-mirrors.yaml /etc/rancher/k3s/registries.yaml"
        
        # Clean up local file
        rm registry-mirrors.yaml
    fi
    
    # Install K3s
    print_progress "Installing K3s..."
    INSTALL_CMD="curl -sfL https://get.k3s.io | INSTALL_K3S_VERSION=$K3S_VERSION K3S_TOKEN=$K3S_TOKEN sh -"
    
    if [ "$SERVER_REGION" = "cn" ]; then
        print_progress "Using China-specific installation options..."
        INSTALL_CMD="curl -sfL https://rancher-mirror.rancher.cn/k3s/k3s-install.sh | INSTALL_K3S_VERSION=$K3S_VERSION K3S_TOKEN=$K3S_TOKEN INSTALL_K3S_MIRROR=cn sh -"
    fi
    
    ssh $SERVER_USER@$SERVER_HOST "$INSTALL_CMD"
    
    # Wait for K3s to start
    print_progress "Waiting for K3s to start..."
    sleep 10
    
    # Check K3s status
    check_status
    
    print_progress "K3s installation completed successfully"
}

# Function to uninstall K3s
uninstall_k3s() {
    print_progress "Uninstalling K3s from server: $SERVER_NAME"
    
    # Check if k3s is installed
    K3S_INSTALLED=$(ssh $SERVER_USER@$SERVER_HOST "command -v k3s > /dev/null && echo 'yes' || echo 'no'")
    
    if [ "$K3S_INSTALLED" != "yes" ]; then
        print_error "K3s is not installed on the server"
        exit 1
    fi
    
    # Confirm uninstallation
    read -p "Are you sure you want to uninstall K3s? This will delete all Kubernetes resources. (y/n): " CONFIRM
    if [ "$CONFIRM" != "y" ]; then
        print_progress "Uninstallation aborted"
        exit 0
    fi
    
    # Uninstall K3s
    print_progress "Uninstalling K3s..."
    ssh $SERVER_USER@$SERVER_HOST "sudo /usr/local/bin/k3s-uninstall.sh"
    
    print_progress "K3s uninstallation completed"
}

# Function to restart K3s
restart_k3s() {
    print_progress "Restarting K3s on server: $SERVER_NAME"
    
    # Check if k3s is installed
    K3S_INSTALLED=$(ssh $SERVER_USER@$SERVER_HOST "command -v k3s > /dev/null && echo 'yes' || echo 'no'")
    
    if [ "$K3S_INSTALLED" != "yes" ]; then
        print_error "K3s is not installed on the server"
        exit 1
    fi
    
    # Get current k3s status
    print_progress "Current k3s service status:"
    ssh $SERVER_USER@$SERVER_HOST "sudo systemctl status k3s | grep Active"
    
    # Restart k3s
    print_progress "Restarting k3s service..."
    ssh $SERVER_USER@$SERVER_HOST "sudo systemctl restart k3s"
    
    # Check if restart was successful
    print_progress "Checking k3s service status after restart:"
    sleep 5
    ssh $SERVER_USER@$SERVER_HOST "sudo systemctl status k3s | grep Active"
    
    print_progress "K3s restarted successfully"
    print_progress "Please wait a few minutes for all pods to stabilize"
}

# Execute the requested command
case $COMMAND in
    install)
        install_k3s
        ;;
    uninstall)
        uninstall_k3s
        ;;
    restart)
        restart_k3s
        ;;
    status)
        check_status
        ;;
    preload-images)
        preload_images
        ;;
    *)
        print_error "Unknown command: $COMMAND"
        show_usage
        ;;
esac

exit 0
