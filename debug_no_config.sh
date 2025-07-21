#!/bin/bash

# Debug script to test LazyOC without kubeconfig

echo "Testing LazyOC without kubeconfig..."
echo "Current ~/.kube/config status:"
ls -la ~/.kube/config 2>/dev/null || echo "No config file found"

echo ""
echo "Running LazyOC in debug mode without alternate screen:"
./bin/lazyoc --debug --no-alt-screen

echo ""
echo "Check lazyoc.log for debug output:"
tail -20 lazyoc.log 2>/dev/null || echo "No log file found"