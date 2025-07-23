#!/bin/bash

# OpenShift Test Applications Setup Script
# This script deploys various test applications to an OpenShift cluster
# for testing LazyOC's pod listing and project management functionality

set -e  # Exit on any error

echo "🚀 Setting up test applications in OpenShift..."
echo "=============================================="

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

# Deploy test applications
echo "🔧 Deploying test applications..."

echo "  📦 Deploying Nginx web server..."
oc new-app --name=nginx-test nginx --as-deployment-config=false > /dev/null 2>&1 || {
    echo "  ⚠️  Nginx deployment failed or already exists"
}

echo "  📦 Deploying Hello World application..."
oc new-app --name=hello-world quay.io/redhattraining/hello-world-nginx:v1.0 --as-deployment-config=false > /dev/null 2>&1 || {
    echo "  ⚠️  Hello World deployment failed or already exists"
}

echo "  📦 Deploying Redis database..."
oc new-app --name=redis-test redis --as-deployment-config=false > /dev/null 2>&1 || {
    echo "  ⚠️  Redis deployment failed or already exists"
}

echo "  📦 Deploying Apache HTTP server..."
oc new-app --name=httpd-test httpd --as-deployment-config=false > /dev/null 2>&1 || {
    echo "  ⚠️  HTTPD deployment failed or already exists"
}

echo ""
echo "⏳ Waiting for deployments to start..."
sleep 5

# Scale some applications for more pods
echo "📈 Scaling applications for better testing..."
echo "  🔄 Scaling nginx to 2 replicas..."
oc scale deployment/nginx-test --replicas=2 > /dev/null 2>&1 || echo "  ⚠️  nginx scaling failed"

echo "  🔄 Scaling hello-world to 3 replicas..."
oc scale deployment/hello-world --replicas=3 > /dev/null 2>&1 || echo "  ⚠️  hello-world scaling failed"

echo ""
echo "⏳ Waiting for pods to be created..."
sleep 10

# Show current pod status
echo "📋 Current pods in project '$CURRENT_PROJECT':"
echo "=============================================="
oc get pods -o wide --no-headers | while read line; do
    echo "  $line"
done

echo ""
echo "📊 Deployment status:"
echo "===================="
oc get deployments --no-headers | while read line; do
    echo "  $line"
done

echo ""
echo "🌐 Services created:"
echo "==================="
oc get services --no-headers | grep -v kubernetes | while read line; do
    echo "  $line"
done

echo ""
echo "✅ Test applications setup complete!"
echo ""
echo "🧪 Testing recommendations:"
echo "  1. Run LazyOC: go build ./cmd/lazyoc && ./lazyoc"
echo "  2. Test pod listing with various statuses"
echo "  3. Test auto-refresh (30s intervals)"
echo "  4. Test manual refresh with 'r' key"
echo "  5. Test project context display (🎯 icon)"
echo "  6. Test Ctrl+P for project switching UI"
echo ""
echo "🧹 When done testing, run: ./scripts/teardown-test-apps.sh"
echo ""