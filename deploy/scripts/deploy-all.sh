#!/bin/bash

# Batch deployment script
set -e

# Read server configuration
SERVER_CONFIG=$(cat deploy/config/servers.yaml)

# Get all server names
SERVERS=$(echo "$SERVER_CONFIG" | yq e '.servers[].name' -)

# Iterate through all servers
for SERVER in $SERVERS; do
    echo "=================================================="
    echo "Deploying to server: $SERVER"
    echo "=================================================="
    
    # Execute single server deployment
    bash deploy/scripts/deploy-server.sh $SERVER
    
    echo ""
done

echo "All servers deployed successfully!"
