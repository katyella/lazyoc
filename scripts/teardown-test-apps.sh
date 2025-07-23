#!/bin/bash

# OpenShift Test Applications Teardown Script
# This script removes all test applications deployed by setup-test-apps.sh
# from the OpenShift cluster

set -e  # Exit on any error

echo "🧹 Tearing down test applications from OpenShift..."
echo "=================================================="

# Check if oc is available and user is logged in
if ! command -v oc &> /dev/null; then
    echo "❌ Error: 'oc' command not found. Please install OpenShift CLI."
    exit 1
fi

# Check if logged in to OpenShift
if ! oc whoami &> /dev/null; then
    echo "❌ Error: Not logged in to OpenShift. Please run 'oc login' first."
    exit 1
fi

# Show current project context
CURRENT_PROJECT=$(oc project -q 2>/dev/null || echo "unknown")
CURRENT_USER=$(oc whoami 2>/dev/null || echo "unknown")
echo "📍 Current project: $CURRENT_PROJECT"
echo "👤 Current user: $CURRENT_USER"
echo ""

# Show what will be deleted
echo "🔍 Applications to be deleted:"
echo "=============================="
APPS=("nginx-test" "hello-world" "redis-test" "httpd-test")

for app in "${APPS[@]}"; do
    if oc get deployment "$app" &> /dev/null; then
        echo "  ✅ $app (found)"
    else
        echo "  ❌ $app (not found)"
    fi
done

echo ""
read -p "❓ Are you sure you want to delete all test applications? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Teardown cancelled."
    exit 0
fi

echo ""
echo "🗑️  Deleting test applications..."

# Delete each application and all its associated resources
for app in "${APPS[@]}"; do
    echo "  🗑️  Deleting $app..."
    
    # Delete all resources with the app label
    oc delete all -l app="$app" --ignore-not-found=true > /dev/null 2>&1 || true
    
    # Also try deleting by deployment name (fallback)
    oc delete deployment "$app" --ignore-not-found=true > /dev/null 2>&1 || true
    oc delete service "$app" --ignore-not-found=true > /dev/null 2>&1 || true
    oc delete route "$app" --ignore-not-found=true > /dev/null 2>&1 || true
    
    echo "    ✅ $app cleanup completed"
done

echo ""
echo "⏳ Waiting for pods to terminate..."
sleep 5

# Clean up any remaining test resources
echo "🧽 Cleaning up remaining test resources..."

# Delete any remaining pods from test apps
for app in "${APPS[@]}"; do
    oc delete pods -l app="$app" --ignore-not-found=true > /dev/null 2>&1 || true
done

# Delete any remaining replica sets
for app in "${APPS[@]}"; do
    oc delete replicaset -l app="$app" --ignore-not-found=true > /dev/null 2>&1 || true
done

echo ""
echo "📋 Remaining resources in project '$CURRENT_PROJECT':"
echo "===================================================="

echo ""
echo "🏃 Pods:"
POD_COUNT=$(oc get pods --no-headers 2>/dev/null | wc -l || echo "0")
if [ "$POD_COUNT" -eq 0 ]; then
    echo "  (no pods found)"
else
    oc get pods --no-headers | while read line; do
        echo "  $line"
    done
fi

echo ""
echo "🚀 Deployments:"
DEPLOY_COUNT=$(oc get deployments --no-headers 2>/dev/null | wc -l || echo "0")
if [ "$DEPLOY_COUNT" -eq 0 ]; then
    echo "  (no deployments found)"
else
    oc get deployments --no-headers | while read line; do
        echo "  $line"
    done
fi

echo ""
echo "🌐 Services:"
SVC_COUNT=$(oc get services --no-headers 2>/dev/null | grep -v kubernetes | wc -l || echo "0")
if [ "$SVC_COUNT" -eq 0 ]; then
    echo "  (no services found, except kubernetes)"
else
    oc get services --no-headers | grep -v kubernetes | while read line; do
        echo "  $line"
    done
fi

echo ""
echo "✅ Test applications teardown complete!"
echo ""
echo "💡 Your OpenShift project is now clean and ready for:"
echo "  - New test deployments"
echo "  - Different LazyOC testing scenarios"
echo "  - Development work"
echo ""
echo "📚 To deploy test apps again, run: ./scripts/setup-test-apps.sh"
echo ""