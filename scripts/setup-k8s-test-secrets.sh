#!/bin/bash

# Kubernetes Test Secrets Setup Script  
# This script creates various test secrets in a Kubernetes cluster
# for testing LazyOC's secret viewing functionality

set -e  # Exit on any error

echo "🔐 Setting up test secrets in Kubernetes..."
echo "==========================================="

# Check if kubectl is available and user is connected
if ! command -v kubectl &> /dev/null; then
    echo "❌ Error: 'kubectl' command not found. Please install Kubernetes CLI."
    exit 1
fi

# Check if connected to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "❌ Error: Not connected to Kubernetes cluster. Please configure kubeconfig."
    exit 1
fi

# Show current context
CURRENT_CONTEXT=$(kubectl config current-context 2>/dev/null || echo "unknown")
CURRENT_NAMESPACE=$(kubectl config view --minify --output 'jsonpath={..namespace}' 2>/dev/null || echo "default")
echo "📍 Current context: $CURRENT_CONTEXT"
echo "📍 Current namespace: $CURRENT_NAMESPACE"
echo ""

echo "🔐 Creating test secrets for manual testing..."

# Database credentials secret
echo "  📦 Creating database credentials secret..."
kubectl delete secret database-credentials --ignore-not-found=true > /dev/null 2>&1
kubectl create secret generic database-credentials \
    --from-literal=username=admin \
    --from-literal=password=super-secret-password-123 \
    --from-literal=host=postgres.example.com \
    --from-literal=port=5432 \
    --from-literal=database=myapp_production > /dev/null 2>&1 || echo "  ⚠️  Database credentials secret creation failed"

# API keys secret
echo "  📦 Creating API keys secret..."
kubectl delete secret api-keys --ignore-not-found=true > /dev/null 2>&1
kubectl create secret generic api-keys \
    --from-literal=stripe-key=sk_test_4eC39HqLyjWDarjtT1zdp7dc \
    --from-literal=aws-access-key=AKIAIOSFODNN7EXAMPLE \
    --from-literal=aws-secret-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
    --from-literal=github-token=ghp_1234567890abcdefghijklmnopqrstuvwxyz \
    --from-literal=slack-webhook=https://hooks.slack.com/services/T00000000/B00000000/XXXXXXXXXXXXXXXXXXXXXXXX > /dev/null 2>&1 || echo "  ⚠️  API keys secret creation failed"

# TLS certificates secret
echo "  📦 Creating TLS certificates secret..."
kubectl delete secret tls-certificates --ignore-not-found=true > /dev/null 2>&1
# Create dummy certificate data (base64 encoded)
kubectl create secret generic tls-certificates \
    --from-literal=tls.crt="$(echo -e '-----BEGIN CERTIFICATE-----\nMIIDXTCCAkWgAwIBAgIJAKoK/heBjcOuMA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV\nBAYTAkFVMRMwEQYDVQQIDApTb21lLVN0YXRlMSEwHwYDVQQKDBhJbnRlcm5ldCBX\naWRnaXRzIFB0eSBMdGQwHhcNMTIwOTEyMjE1MjAyWhcNMTUwOTEyMjE1MjAyWjBF\nMQswCQYDVQQGEwJBVTETMBEGA1UECAwKU29tZS1TdGF0ZTEhMB8GA1UECgwYSW50\nZXJuZXQgV2lkZ2l0cyBQdHkgTHRkMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIB\nCgKCAQEAwUdHPiQmXwQs04XmhWmJKnTBAu6wjTQRPM93fJkIDz3jZbHJqD9BcXFz\nZ7dEACbLyJk13V5dGKJYCHfY2DjL+m3j2F5GhWoO3hKw1ZcLIYmvZ5W2Q3R6gVA1\n2P7YvW8JaCGCF+3jM8hAF0JBHWvUk0F8s9Ev4hKQ5Sv/HxWfvGgYlQqiWK3b3Y9Z\nM4fk5m6Bz+xMqW+6DthgL4Q5iWrCUBB2o5yRn5d9k6BG4v5K2iRgL3lT8gA7wKgZ\nT1iGX1k8K+T4d7b9Q9w3kP9aCfU1K4X8RlVHdPeJ+6r8T8A5z2n6V2hD8C1v9mD+\nK4I+4t/qywIDAQABo1AwTjAdBgNVHQ4EFgQUhKs61e4zPx0PSWaFqFqmAFGj9g0w\nHwYDVR0jBBgwFoAUhKs61e4zPx0PSWaFqFqmAFGj9g0wDAYDVR0TBAUwAwEB/zAN\nBgkqhkiG9w0BAQUFAAOCAQEAGcnqtZNlNqFH1BKpKr7Ky8k/wZ8Y3Fh3i2d7B4h7\n-----END CERTIFICATE-----')" \
    --from-literal=tls.key="$(echo -e '-----BEGIN PRIVATE KEY-----\nMIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDBR0c+JCZfBCzT\nheaFaYkqdMEC7rCNNBE8z3d8mQgPPeNlscmoP0FxcXNnt0QAJsvImTXdXl0Yolg\nId9jYOMv6bePYXkaFag7eErDVlwshia9nlbZDdHqBUDXY/ti9bwloIYIX7eMzw\nEAXQkEda9STQXyz0S/iEpDlK/8fFZ+8aBiVCqJYrdvdj1kzh+TmboHP7Eypb7\noO2GAvhDmJasJQEHajnJGfl32ToEbi/kraJGAveVPyADvAqBlPWIZfWTwr5Ph\n3tv1D3DeQ/1oJ9TUrhfxGVUd094n7qvxPwDnPafpXaEPwLW/2YP4rgj7i3+rL\nAgMBAAECggEBALVGvz7Ql0VGKz90n2vXdmm7Y9Tk+c2fIqp3zAjQxGnK4vE1iK\nfKZ7TlCG3z5r7fL1hY3bLr+e2K2oI8u2xLz3mXc6W4V4n9m2H+K5iB8jJE9Q\n3K9j7F4hG2Q9oE8mT8zKXjL1l5X3kYL3uKvVJF8gS7ZnF1t7fF+F5oZ6j9d4\nQ8q8zVcXhL2U5fvJx9oKjW8Tn2L5qR7fE5Y8hP5rN3+xV4c3k6bQ1A1f5Y4k\n2cY5N9z4K6r8rGo4d4U7M5+L9K3fJ8z5W4hV7b3S8oI5u7yT4f7Kj2V9t8pL\n5fR4oX2k9Y7E1+c8z6bU9bJ4q7d5W8Vy8pJyFt7oECgYEA46j8D1p8L7q2fI\n-----END PRIVATE KEY-----')" > /dev/null 2>&1 || echo "  ⚠️  TLS certificates secret creation failed"

# Application configuration secret
echo "  📦 Creating application configuration secret..."
kubectl delete secret app-config --ignore-not-found=true > /dev/null 2>&1
kubectl create secret generic app-config \
    --from-literal=debug=true \
    --from-literal=log-level=INFO \
    --from-literal=max-connections=100 \
    --from-literal=timeout=30s \
    --from-literal=cache-ttl=3600 \
    --from-literal=feature-flags="{\"new-ui\": true, \"beta-api\": false, \"analytics\": true}" \
    --from-literal=environment=development > /dev/null 2>&1 || echo "  ⚠️  Application configuration secret creation failed"

# Docker registry credentials secret
echo "  📦 Creating Docker registry credentials secret..."
kubectl delete secret docker-registry-creds --ignore-not-found=true > /dev/null 2>&1
kubectl create secret generic docker-registry-creds \
    --from-literal=registry=docker.io \
    --from-literal=username=mycompany \
    --from-literal=password=my-registry-password-456 \
    --from-literal=email=devops@company.com > /dev/null 2>&1 || echo "  ⚠️  Docker registry credentials secret creation failed"

# OAuth tokens secret  
echo "  📦 Creating OAuth tokens secret..."
kubectl delete secret oauth-tokens --ignore-not-found=true > /dev/null 2>&1
kubectl create secret generic oauth-tokens \
    --from-literal=google-oauth-client-id=123456789012-abcdefghijklmnopqrstuvwxyz123456.apps.googleusercontent.com \
    --from-literal=google-oauth-client-secret=GOCSPX-abcdefghijklmnopqrstuvwxyz123456 \
    --from-literal=github-oauth-client-id=Iv1.abcdefghijklmnop \
    --from-literal=github-oauth-client-secret=abcdefghijklmnopqrstuvwxyz1234567890abcdef \
    --from-literal=jwt-secret=my-super-secret-jwt-signing-key-that-should-be-very-long > /dev/null 2>&1 || echo "  ⚠️  OAuth tokens secret creation failed"

# Simple credentials secret (for basic testing)
echo "  📦 Creating simple credentials secret..."
kubectl delete secret simple-creds --ignore-not-found=true > /dev/null 2>&1
kubectl create secret generic simple-creds \
    --from-literal=user=admin \
    --from-literal=pass=password123 > /dev/null 2>&1 || echo "  ⚠️  Simple credentials secret creation failed"

# Multi-line configuration secret
echo "  📦 Creating multi-line config secret..."
kubectl delete secret multiline-config --ignore-not-found=true > /dev/null 2>&1
kubectl create secret generic multiline-config \
    --from-literal=nginx.conf="$(cat << 'EOF'
server {
    listen 80;
    server_name example.com;
    
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
    
    location /health {
        return 200 "OK";
        add_header Content-Type text/plain;
    }
}
EOF
)" \
    --from-literal=app.yaml="$(cat << 'EOF'
apiVersion: v1
kind: ConfigMap
metadata:
  name: example-config
data:
  database_url: postgres://user:pass@localhost/db
  redis_url: redis://localhost:6379/0
  log_level: debug
  features:
    - feature_a
    - feature_b
    - feature_c
EOF
)" > /dev/null 2>&1 || echo "  ⚠️  Multi-line config secret creation failed"

echo ""
echo "🔐 Secrets created:"
echo "=================="
kubectl get secrets --no-headers | grep -v default-token | while read line; do
    echo "  $line"
done

echo ""
echo "✅ Test secrets setup complete!"
echo ""
echo "🧪 Testing recommendations:"
echo "  1. Run LazyOC: go build -o ./bin/lazyoc ./cmd/lazyoc && ./bin/lazyoc"
echo "  2. Navigate to Secrets tab"
echo "  3. 🔐 Test secret viewing: Select a secret → Press Enter"
echo "     - Use j/k to navigate between secret keys"
echo "     - Press 'm' to toggle masking/unmasking of values"
echo "     - Press 'c' to copy selected key's value to clipboard"
echo "     - Press 'C' to copy entire secret as JSON to clipboard"
echo "     - Press 'esc' or 'q' to close the modal"
echo ""
echo "🧹 When done testing, run:"
echo "   kubectl delete secret database-credentials api-keys tls-certificates app-config docker-registry-creds oauth-tokens simple-creds multiline-config"
echo ""