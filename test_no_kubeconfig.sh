#!/bin/bash

echo "Testing LazyOC without kubeconfig..."

# Backup existing kubeconfig if it exists
if [ -f ~/.kube/config ]; then
    echo "Backing up existing kubeconfig..."
    mv ~/.kube/config ~/.kube/config.test_backup
fi

echo ""
echo "Starting LazyOC - you should see:"
echo "1. Main panel: 'Not connected to any cluster' message"
echo "2. Log panel: Warning about no kubeconfig found"
echo "3. Header: Disconnected status"
echo ""
echo "Starting in 2 seconds..."
sleep 2

./bin/lazyoc

# Restore kubeconfig if we backed it up
if [ -f ~/.kube/config.test_backup ]; then
    echo ""
    echo "Restoring kubeconfig..."
    mv ~/.kube/config.test_backup ~/.kube/config
fi

echo "Test complete!"