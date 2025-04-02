#!/bin/bash

# Script to clean up deployment scripts
# This script removes redundant deployment scripts and keeps only the essential ones
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

# Scripts to keep
KEEP_SCRIPTS=(
    "deploy-k3s.sh"        # Main deployment script
    "check-k3s-status.sh"  # Status checking script
    "cleanup-scripts.sh"   # This script
)

print_progress "Starting deployment scripts cleanup..."

# Get all scripts in the directory
cd "$(dirname "$0")"
ALL_SCRIPTS=(*.sh)

# Count scripts
TOTAL_SCRIPTS=${#ALL_SCRIPTS[@]}
KEPT_SCRIPTS=0
REMOVED_SCRIPTS=0

# Process each script
for script in "${ALL_SCRIPTS[@]}"; do
    keep=false
    
    # Check if script should be kept
    for keep_script in "${KEEP_SCRIPTS[@]}"; do
        if [ "$script" = "$keep_script" ]; then
            keep=true
            break
        fi
    done
    
    if [ "$keep" = true ]; then
        print_progress "Keeping script: $script"
        KEPT_SCRIPTS=$((KEPT_SCRIPTS + 1))
    else
        print_warning "Removing script: $script"
        rm "$script"
        REMOVED_SCRIPTS=$((REMOVED_SCRIPTS + 1))
    fi
done

print_progress "Cleanup completed:"
print_progress "- Total scripts: $TOTAL_SCRIPTS"
print_progress "- Kept scripts: $KEPT_SCRIPTS"
print_progress "- Removed scripts: $REMOVED_SCRIPTS"

print_progress "The deployment directory now contains only the essential scripts."
