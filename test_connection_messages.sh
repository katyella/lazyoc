#!/bin/bash

echo "=== Testing LazyOC Connection Messages ==="
echo ""

# Test 1: Without kubeconfig
echo "Test 1: Running without kubeconfig"
echo "--------------------------------"
mv ~/.kube/config ~/.kube/config.test_backup 2>/dev/null || true
echo "Temporarily moved kubeconfig away"
echo ""
echo "Starting LazyOC - you should see:"
echo "- Header: 'Not connected - Run oc login or use --kubeconfig'"
echo "- Main panel: Instructions on how to connect"
echo "- Log panel: Warning about no kubeconfig found"
echo ""
echo "Press 'q' to quit when ready..."
./bin/lazyoc

echo ""
echo "Test 2: With kubeconfig restored"
echo "-------------------------------"
mv ~/.kube/config.test_backup ~/.kube/config 2>/dev/null || true
echo "Restored kubeconfig"
echo ""
echo "Starting LazyOC - should attempt to connect automatically"
echo "Press 'q' to quit when ready..."
./bin/lazyoc

echo ""
echo "=== Test Complete ==="