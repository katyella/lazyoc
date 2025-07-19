#!/bin/bash

# test-tui.sh - Script to manually test the TUI functionality

set -e

echo "Building LazyOC..."
make build

echo ""
echo "Testing TUI functionality..."
echo "================================"
echo ""

echo "1. Testing version flag:"
./bin/lazyoc --version
echo ""

echo "2. Testing help flag:"
./bin/lazyoc --help | head -10
echo ""

echo "3. Testing debug mode (will create lazyoc.log):"
echo "   Run: ./bin/lazyoc --debug"
echo "   Then press 'q' to quit and check lazyoc.log"
echo ""

echo "4. Manual TUI test instructions:"
echo "   Run: ./bin/lazyoc"
echo "   Try these keys:"
echo "   - Tab/Shift+Tab or h/l: Navigate between tabs"
echo "   - ?: Toggle help overlay"
echo "   - Ctrl+D: Toggle debug mode"
echo "   - q or Ctrl+C: Quit"
echo ""

echo "5. Test different modes:"
echo "   ./bin/lazyoc --no-alt-screen    # Disable alternate screen"
echo "   ./bin/lazyoc --debug            # Enable debug logging"
echo ""

echo "Ready to test! Run the commands above to verify TUI functionality."