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

echo "  📦 Deploying continuous logging application..."
# Delete existing log-generator if it exists to ensure clean deployment
oc delete deployment log-generator --ignore-not-found=true > /dev/null 2>&1

# Create the deployment with proper logging command from the start
cat <<EOF | oc apply -f - > /dev/null 2>&1 || echo "  ⚠️  Log generator deployment failed"
apiVersion: apps/v1
kind: Deployment
metadata:
  name: log-generator
  labels:
    app: log-generator
spec:
  replicas: 1
  selector:
    matchLabels:
      app: log-generator
  template:
    metadata:
      labels:
        app: log-generator
    spec:
      containers:
      - name: log-generator
        image: alpine:latest
        command: ["/bin/sh"]
        args: ["-c", "counter=1; while true; do echo \"\$(date) [INFO] Log entry #\$counter - Application is running smoothly\"; echo \"\$(date) [DEBUG] Processing request ID: \$((counter * 7 % 9999))\"; echo \"\$(date) [WARN] Memory usage at \$((counter % 40 + 60))%\"; if [ \$((counter % 10)) -eq 0 ]; then echo \"\$(date) [ERROR] Simulated error condition detected\"; fi; counter=\$((counter + 1)); sleep 5; done"]
EOF

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
echo "🔐 Creating test secrets for manual testing..."

# Database credentials secret
echo "  📦 Creating database credentials secret..."
oc delete secret database-credentials --ignore-not-found=true > /dev/null 2>&1
oc create secret generic database-credentials \
    --from-literal=username=admin \
    --from-literal=password=super-secret-password-123 \
    --from-literal=host=postgres.example.com \
    --from-literal=port=5432 \
    --from-literal=database=myapp_production > /dev/null 2>&1 || echo "  ⚠️  Database credentials secret creation failed"

# API keys secret
echo "  📦 Creating API keys secret..."
oc delete secret api-keys --ignore-not-found=true > /dev/null 2>&1
oc create secret generic api-keys \
    --from-literal=stripe-key=sk_test_4eC39HqLyjWDarjtT1zdp7dc \
    --from-literal=aws-access-key=AKIAIOSFODNN7EXAMPLE \
    --from-literal=aws-secret-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
    --from-literal=github-token=ghp_1234567890abcdefghijklmnopqrstuvwxyz \
    --from-literal=slack-webhook=https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX > /dev/null 2>&1 || echo "  ⚠️  API keys secret creation failed"

# TLS certificates secret
echo "  📦 Creating TLS certificates secret..."
oc delete secret tls-certificates --ignore-not-found=true > /dev/null 2>&1
# Create dummy certificate data (base64 encoded)
oc create secret generic tls-certificates \
    --from-literal=tls.crt="$(echo -e '-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV\nBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX\naWRnaXRzIFB0eSBMdGQwHhcNMTIwOTEyMjE1MjAyWhcNMTUwOTEyMjE1MjAyWjBF\nMQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50\nZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\nCgKCAQEAwUdHPiQmXwQs04XmhWmJKnTBAu6wjTQRPM93fJkIDz3jZbHJqD9BcXFz\nZ7dEACbLyJk13V5dGKJYCHfY2DjL+m3j2F5GhWoO3hKw1ZcLIYmvZ5W2Q3R6gVA1\n2P7YvW8JaCGCF+3jM8hAF0JBHWvUk0F8s9Ev4hKQ5Sv/HxWfvGgYlQqiWK3b3Y9Z\nM4fk5m6Bz+xMqW+6DthgL4Q5iWrCUBB2o5yRn5d9k6BG4v5K2iRgL3lT8gA7wKgZ\nT1iGX1k8K+T4d7b9Q9w3kP9aCfU1K4X8RlVHdPeJ+6r8T8A5z2n6V2hD8C1v9mD+\nK4I+4t/qywIDAQABo1AwTjAdBgNVHQ4EFgQUhKs61e4zPx0PSWaFqFqmAFGj9g0w\nHwYDVR0jBBgwFoAUhKs61e4zPx0PSWaFqFqmAFGj9g0wDAYDVR0TBAUwAwEB/zAN\nBgkqhkiG9w0BAQUFAAOCAQEAGcnqtZNlNqFH1BKpKr7Ky8k/wZ8Y3Fh3i2d7B4h7\n-----END CERTIFICATE-----' | base64 -w 0)" \
    --from-literal=tls.key="$(echo -e '-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDBR0c+JCZfBCzT\nheaFaYkqdMEC7rCNNBE8z3d8mQgPPeNlscmoP0FxcXNnt0QAJsvImTXdXl0Yolg\nId9jYOMv6bePYXkaFag7eErDVlwshia9nlbZDdHqBUDXY/ti9bwloIYIX7eMzw\nEAXQkEda9STQXyz0S/iEpDlK/8fFZ+8aBiVCqJYrdvdj1kzh+TmboHP7Eypb7\noO2GAvhDmJasJQEHajnJGfl32ToEbi/kraJGAveVPyADvAqBlPWIZfWTwr5Ph\n3tv1D3DeQ/1oJ9TUrhfxGVUd094n7qvxPwDnPafpXaEPwLW/2YP4rgj7i3+rL\nAgMBAAECggEBALVGvz7Ql0VGKz90n2vXdmm7Y9Tk+c2fIqp3zAjQxGnK4vE1iK\nfKZ7TlCG3z5r7fL1hY3bLr+e2K2oI8u2xLz3mXc6W4V4n9m2H+K5iB8jJE9Q\n3K9j7F4hG2Q9oE8mT8zKXjL1l5X3kYL3uKvVJF8gS7ZnF1t7fF+F5oZ6j9d4\nQ8q8zVcXhL2U5fvJx9oKjW8Tn2L5qR7fE5Y8hP5rN3+xV4c3k6bQ1A1f5Y4k\n2cY5N9z4K6r8rGo4d4U7M5+L9K3fJ8z5W4hV7b3S8oI5u7yT4f7Kj2V9t8pL\n5fR4oX2k9Y7E1+c8z6bU9bJ4q7d5W8Vy8pJyFt7oECgYEA46j8D1p8L7q2fI\n-----END PRIVATE KEY-----' | base64 -w 0)" > /dev/null 2>&1 || echo "  ⚠️  TLS certificates secret creation failed"

# Application configuration secret
echo "  📦 Creating application configuration secret..."
oc delete secret app-config --ignore-not-found=true > /dev/null 2>&1
oc create secret generic app-config \
    --from-literal=debug=true \
    --from-literal=log-level=INFO \
    --from-literal=max-connections=100 \
    --from-literal=timeout=30s \
    --from-literal=cache-ttl=3600 \
    --from-literal=feature-flags="{\"new-ui\": true, \"beta-api\": false, \"analytics\": true}" \
    --from-literal=environment=development > /dev/null 2>&1 || echo "  ⚠️  Application configuration secret creation failed"

# Docker registry credentials secret
echo "  📦 Creating Docker registry credentials secret..."
oc delete secret docker-registry-creds --ignore-not-found=true > /dev/null 2>&1
oc create secret generic docker-registry-creds \
    --from-literal=registry=docker.io \
    --from-literal=username=mycompany \
    --from-literal=password=my-registry-password-456 \
    --from-literal=email=devops@company.com > /dev/null 2>&1 || echo "  ⚠️  Docker registry credentials secret creation failed"

# OAuth tokens secret  
echo "  📦 Creating OAuth tokens secret..."
oc delete secret oauth-tokens --ignore-not-found=true > /dev/null 2>&1
oc create secret generic oauth-tokens \
    --from-literal=google-oauth-client-id=123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com \
    --from-literal=google-oauth-client-secret=GOCSPX-abcdefghijklmnopqrstuvwxyz123456 \
    --from-literal=github-oauth-client-id=Iv1.abcdefghijklmnop \
    --from-literal=github-oauth-client-secret=abcdefghijklmnopqrstuvwxyz1234567890abcdef \
    --from-literal=jwt-secret=my-super-secret-jwt-signing-key-that-should-be-very-long > /dev/null 2>&1 || echo "  ⚠️  OAuth tokens secret creation failed"

# Simple credentials secret (for basic testing)
echo "  📦 Creating simple credentials secret..."
oc delete secret simple-creds --ignore-not-found=true > /dev/null 2>&1
oc create secret generic simple-creds \
    --from-literal=user=admin \
    --from-literal=pass=password123 > /dev/null 2>&1 || echo "  ⚠️  Simple credentials secret creation failed"

echo ""
echo "🔐 Secrets created:"
echo "=================="
oc get secrets --no-headers | grep -v default-token | while read line; do
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
echo "  7. Test log viewing with the log-generator pod (continuously logging)"
echo "  8. Toggle between app logs and pod logs with 'l' key"
echo "  9. 🔐 Test secret viewing: Navigate to Secrets tab → Select a secret → Press Enter"
echo "     - Use j/k to navigate between secret keys"
echo "     - Press 'm' to toggle masking/unmasking of values"
echo "     - Press 'c' to copy selected key's value to clipboard"
echo "     - Press 'C' to copy entire secret as JSON to clipboard"
echo "     - Press 'esc' or 'q' to close the modal"
echo "  10. Test service-level logs: Navigate to Services tab → Select a service → Press 'L'"
echo ""
echo "🧹 When done testing, run: ./scripts/teardown-test-apps.sh"
echo ""
echo "🚀 RECOMMENDED: For a comprehensive LazyOC demonstration:"
echo "  📦 Use ./scripts/setup-production-demo.sh instead"
echo "  🏪 Deploys Google Cloud's 'Online Boutique' - 11 production microservices"
echo "  ⭐ Perfect for showcasing ALL LazyOC features with real applications"
echo ""
echo "📝 Alternative options:"
echo "  - ./scripts/setup-k8s-test-secrets.sh (Kubernetes-only secrets)"
echo "  - This script (simple test apps - legacy)"
echo ""