#!/bin/bash

# LazyOC Production Demo Environment Teardown
# Removes Google Cloud Platform's "Online Boutique" microservices demo

set -e  # Exit on any error

echo "🧹 Tearing down production LazyOC demo environment..."
echo "===================================================="

# Check if kubectl/oc is available
if command -v oc &> /dev/null; then
    CLI_TOOL="oc"
    echo "✅ Using OpenShift CLI (oc)"
elif command -v kubectl &> /dev/null; then
    CLI_TOOL="kubectl"
    echo "✅ Using Kubernetes CLI (kubectl)"
else
    echo "❌ Error: Neither 'oc' nor 'kubectl' command found."
    exit 1
fi

# Check if connected
if ! $CLI_TOOL cluster-info &> /dev/null; then
    echo "❌ Error: Not connected to a cluster."
    exit 1
fi

# Show current context
if [ "$CLI_TOOL" = "oc" ]; then
    CURRENT_PROJECT=$($CLI_TOOL project -q 2>/dev/null || echo "unknown")
    echo "📍 Current project: $CURRENT_PROJECT"
else
    CURRENT_CONTEXT=$($CLI_TOOL config current-context 2>/dev/null || echo "unknown")
    CURRENT_NAMESPACE=$($CLI_TOOL config view --minify --output 'jsonpath={..namespace}' 2>/dev/null || echo "default")
    echo "📍 Current context: $CURRENT_CONTEXT"
    echo "📍 Current namespace: $CURRENT_NAMESPACE"
fi

echo ""
echo "🔍 Resources to be deleted:"
echo "=========================="

# Check what exists
BOUTIQUE_PODS=$($CLI_TOOL get pods -l app=online-boutique --no-headers 2>/dev/null | wc -l || echo "0")
DEMO_SECRETS=$($CLI_TOOL get secrets -l demo=lazyoc-production --no-headers 2>/dev/null | wc -l || echo "0")
DEMO_CONFIGMAPS=$($CLI_TOOL get configmaps -l demo=lazyoc-production --no-headers 2>/dev/null | wc -l || echo "0")

echo "  📦 Online Boutique pods: $BOUTIQUE_PODS"
echo "  🔐 Demo secrets: $DEMO_SECRETS" 
echo "  📋 Demo configmaps: $DEMO_CONFIGMAPS"

if [ "$CLI_TOOL" = "oc" ] && oc api-resources | grep -q "routes.*route.openshift.io"; then
    DEMO_ROUTES=$($CLI_TOOL get routes -l demo=lazyoc-production --no-headers 2>/dev/null | wc -l || echo "0")
    echo "  🌍 Demo routes: $DEMO_ROUTES"
fi

TOTAL_RESOURCES=$((BOUTIQUE_PODS + DEMO_SECRETS + DEMO_CONFIGMAPS))
if [ $TOTAL_RESOURCES -eq 0 ]; then
    echo ""
    echo "✅ No demo resources found. Environment is already clean."
    exit 0
fi

echo ""
read -p "❓ Are you sure you want to delete all demo resources? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Teardown cancelled."
    exit 0
fi

echo ""
echo "🗑️  Deleting Online Boutique application..."

# Delete the main application using the official manifest
echo "  📦 Removing Online Boutique microservices..."
if curl -s -f https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/v0.6.0/release/kubernetes-manifests.yaml > /dev/null; then
    $CLI_TOOL delete -f https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/v0.6.0/release/kubernetes-manifests.yaml --ignore-not-found=true > /dev/null 2>&1 || true
else
    # Fallback: delete by labels
    echo "  ⚠️  Could not fetch manifest, using label-based deletion..."
    $CLI_TOOL delete all -l app=online-boutique --ignore-not-found=true > /dev/null 2>&1 || true
fi

echo "  🔐 Removing demo secrets..."
$CLI_TOOL delete secrets -l demo=lazyoc-production --ignore-not-found=true > /dev/null 2>&1 || true

echo "  📋 Removing demo configmaps..."
$CLI_TOOL delete configmaps -l demo=lazyoc-production --ignore-not-found=true > /dev/null 2>&1 || true

if [ "$CLI_TOOL" = "oc" ] && oc api-resources | grep -q "routes.*route.openshift.io"; then
    echo "  🌍 Removing demo routes..."
    $CLI_TOOL delete routes -l demo=lazyoc-production --ignore-not-found=true > /dev/null 2>&1 || true
fi

echo ""
echo "⏳ Waiting for resources to be terminated..."
sleep 5

echo ""
echo "📋 Remaining resources summary:"
echo "==============================="

echo ""
echo "🏃 Pods:"
REMAINING_PODS=$($CLI_TOOL get pods -l app=online-boutique --no-headers 2>/dev/null | wc -l || echo "0")
if [ "$REMAINING_PODS" -eq 0 ]; then
    echo "  ✅ No Online Boutique pods found"
else
    echo "  ⏳ $REMAINING_PODS pods still terminating..."
fi

echo ""
echo "🚀 Deployments:"
REMAINING_DEPLOYMENTS=$($CLI_TOOL get deployments -l app=online-boutique --no-headers 2>/dev/null | wc -l || echo "0")
if [ "$REMAINING_DEPLOYMENTS" -eq 0 ]; then
    echo "  ✅ No Online Boutique deployments found"
else
    echo "  ⏳ $REMAINING_DEPLOYMENTS deployments still terminating..."
fi

echo ""
echo "🌐 Services:"
REMAINING_SERVICES=$($CLI_TOOL get services -l app=online-boutique --no-headers 2>/dev/null | wc -l || echo "0")
if [ "$REMAINING_SERVICES" -eq 0 ]; then
    echo "  ✅ No Online Boutique services found"
else
    echo "  ⏳ $REMAINING_SERVICES services still terminating..."
fi

echo ""
echo "🔐 Demo Secrets:"
REMAINING_SECRETS=$($CLI_TOOL get secrets -l demo=lazyoc-production --no-headers 2>/dev/null | wc -l || echo "0")
if [ "$REMAINING_SECRETS" -eq 0 ]; then
    echo "  ✅ No demo secrets found"
else
    echo "  ⏳ $REMAINING_SECRETS secrets still exist"
fi

echo ""
echo "📋 Demo ConfigMaps:"
REMAINING_CONFIGMAPS=$($CLI_TOOL get configmaps -l demo=lazyoc-production --no-headers 2>/dev/null | wc -l || echo "0")
if [ "$REMAINING_CONFIGMAPS" -eq 0 ]; then
    echo "  ✅ No demo configmaps found"
else
    echo "  ⏳ $REMAINING_CONFIGMAPS configmaps still exist"
fi

echo ""
echo "✅ Production demo teardown complete!"
echo ""
echo "💡 Your cluster is now clean and ready for:"
echo "  - New demo deployments"
echo "  - Different LazyOC testing scenarios"  
echo "  - Development work"
echo ""
echo "📚 To deploy the demo again, run: ./scripts/setup-production-demo.sh"
echo ""