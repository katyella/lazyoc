#!/bin/bash

# test_tui.sh - Test script for TUI with K8s integration

set -e

echo "Building LazyOC..."
go build -o ./bin/lazyoc ./cmd/lazyoc

echo ""
echo "LazyOC built successfully!"
echo ""
echo "Testing connection modes:"
echo ""
echo "1. Test with default kubeconfig location:"
echo "   ./bin/lazyoc"
echo ""
echo "2. Test with specific kubeconfig:"
echo "   ./bin/lazyoc --kubeconfig ~/.kube/config"
echo ""
echo "3. Test with debug mode:"
echo "   ./bin/lazyoc --debug"
echo ""
echo "4. Test help:"
echo "   ./bin/lazyoc --help"
echo ""
echo "Key bindings:"
echo "- j/k: Navigate pod list"
echo "- r: Refresh pod list"
echo "- tab: Switch panels"
echo "- d: Toggle details panel"
echo "- L: Toggle logs panel"
echo "- ?: Show help"
echo "- q: Quit"
echo ""
echo "Run one of the commands above to test the TUI!"