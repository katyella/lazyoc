#!/bin/bash

# LazyOC Production Demo Environment Setup
# Deploys Google Cloud Platform's "Online Boutique" microservices demo
# A production-ready e-commerce application with 11 microservices

set -e  # Exit on any error

echo "🚀 Setting up production-ready LazyOC demo environment..."
echo "======================================================="
echo ""
echo "This will deploy Google Cloud Platform's 'Online Boutique' demo:"
echo "  🏪 A complete e-commerce application with 11 microservices"
echo "  📦 Frontend, Backend, Database, Cache, and more services"
echo "  🔧 Production-ready containerized applications"
echo "  🌐 Real gRPC communication between services"
echo "  📊 Comprehensive observability and monitoring"
echo "  🎯 Perfect for demonstrating ALL LazyOC capabilities"
echo ""

# Check if kubectl/oc is available
if command -v oc &> /dev/null; then
    CLI_TOOL="oc"
    echo "✅ Using OpenShift CLI (oc)"
elif command -v kubectl &> /dev/null; then
    CLI_TOOL="kubectl"
    echo "✅ Using Kubernetes CLI (kubectl)"
else
    echo "❌ Error: Neither 'oc' nor 'kubectl' command found."
    echo "   Please install OpenShift CLI or Kubernetes CLI."
    exit 1
fi

# Check if logged in / connected
if [ "$CLI_TOOL" = "oc" ]; then
    if ! oc whoami &> /dev/null; then
        echo "❌ Error: Not logged in to OpenShift."
        echo "   Please run 'oc login' first."
        exit 1
    fi
else
    if ! kubectl auth can-i get pods &> /dev/null; then
        echo "❌ Error: Not connected to Kubernetes or insufficient permissions."
        echo "   Please configure kubeconfig first."
        exit 1
    fi
fi

# Show current context
if [ "$CLI_TOOL" = "oc" ]; then
    CURRENT_PROJECT=$($CLI_TOOL project -q 2>/dev/null || echo "unknown")
    CURRENT_USER=$($CLI_TOOL whoami 2>/dev/null || echo "unknown")
    echo "📍 Current project: $CURRENT_PROJECT"
    echo "👤 Current user: $CURRENT_USER"
else
    CURRENT_CONTEXT=$($CLI_TOOL config current-context 2>/dev/null || echo "unknown")
    CURRENT_NAMESPACE=$($CLI_TOOL config view --minify --output 'jsonpath={..namespace}' 2>/dev/null || echo "default")
    echo "📍 Current context: $CURRENT_CONTEXT"
    echo "📍 Current namespace: $CURRENT_NAMESPACE"
fi

echo ""
read -p "❓ Continue with production demo setup? (y/N): " -n 1 -r
echo ""
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Setup cancelled."
    exit 0
fi

echo ""
echo "🧹 Cleaning up any existing demo resources..."
$CLI_TOOL delete deployment,service,secret,configmap emailservice checkoutservice recommendationservice frontend paymentservice productcatalogservice cartservice currencyservice shippingservice redis-cart adservice log-generator --ignore-not-found=true > /dev/null 2>&1 || true
$CLI_TOOL delete secrets,configmaps -l demo=lazyoc-production --ignore-not-found=true > /dev/null 2>&1 || true

echo ""
echo "📦 Phase 1: Downloading Google Cloud Online Boutique"
echo "===================================================="

# Create temporary directory
TEMP_DIR=$(mktemp -d)
cd "$TEMP_DIR"

echo "  📥 Downloading Online Boutique v0.6.0..."
if command -v git &> /dev/null; then
    git clone --depth 1 --branch v0.6.0 https://github.com/GoogleCloudPlatform/microservices-demo.git > /dev/null 2>&1
    cd microservices-demo
else
    # Fallback: download specific release files
    echo "  ⚠️  Git not found, downloading release files directly..."
    curl -s -L -o kubernetes-manifests.yaml https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/v0.6.0/release/kubernetes-manifests.yaml
fi

echo ""
echo "🚀 Phase 2: Deploying Online Boutique Microservices"
echo "==================================================="

if [ -f "release/kubernetes-manifests.yaml" ]; then
    MANIFEST_FILE="release/kubernetes-manifests.yaml"
elif [ -f "kubernetes-manifests.yaml" ]; then
    MANIFEST_FILE="kubernetes-manifests.yaml"
else
    echo "❌ Error: Could not find Kubernetes manifests file."
    exit 1
fi

echo "  📦 Deploying all microservices..."

# Check if we can create LoadBalancer services (quota check)
if $CLI_TOOL auth can-i create services --subresource=status 2>/dev/null; then
    echo "  ✅ Full permissions detected, deploying original manifest..."
    $CLI_TOOL apply -f "$MANIFEST_FILE"
else
    echo "  ⚠️  Limited permissions detected, creating quota-friendly version..."
    
    # Create a quota-friendly version without LoadBalancer services
    cat > boutique-quota-friendly.yaml << 'EOF'
# Modified Online Boutique for quota-limited environments
# Removes LoadBalancer services and reduces resource requests

apiVersion: apps/v1
kind: Deployment
metadata:
  name: emailservice
  labels:
    app: online-boutique
    service: emailservice
spec:
  selector:
    matchLabels:
      app: emailservice
  template:
    metadata:
      labels:
        app: emailservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/emailservice:v0.6.0
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        resources:
          requests:
            cpu: 50m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: emailservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: emailservice
  ports:
  - name: grpc
    port: 5000
    targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: checkoutservice
  labels:
    app: online-boutique
    service: checkoutservice
spec:
  selector:
    matchLabels:
      app: checkoutservice
  template:
    metadata:
      labels:
        app: checkoutservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/checkoutservice:v0.6.0
        ports:
        - containerPort: 5050
        env:
        - name: PORT
          value: "5050"
        - name: PRODUCT_CATALOG_SERVICE_ADDR
          value: "productcatalogservice:3550"
        - name: SHIPPING_SERVICE_ADDR
          value: "shippingservice:50051"
        - name: PAYMENT_SERVICE_ADDR
          value: "paymentservice:50051"
        - name: EMAIL_SERVICE_ADDR
          value: "emailservice:5000"
        - name: CURRENCY_SERVICE_ADDR
          value: "currencyservice:7000"
        - name: CART_SERVICE_ADDR
          value: "cartservice:7070"
        resources:
          requests:
            cpu: 50m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: checkoutservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: checkoutservice
  ports:
  - name: grpc
    port: 5050
    targetPort: 5050
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: recommendationservice
  labels:
    app: online-boutique
    service: recommendationservice
spec:
  selector:
    matchLabels:
      app: recommendationservice
  template:
    metadata:
      labels:
        app: recommendationservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/recommendationservice:v0.6.0
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: PRODUCT_CATALOG_SERVICE_ADDR
          value: "productcatalogservice:3550"
        resources:
          requests:
            cpu: 50m
            memory: 110Mi
          limits:
            cpu: 100m
            memory: 220Mi
---
apiVersion: v1
kind: Service
metadata:
  name: recommendationservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: recommendationservice
  ports:
  - name: grpc
    port: 8080
    targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend
  labels:
    app: online-boutique
    service: frontend
spec:
  selector:
    matchLabels:
      app: frontend
  template:
    metadata:
      labels:
        app: frontend
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/frontend:v0.6.0
        ports:
        - containerPort: 8080
        env:
        - name: PORT
          value: "8080"
        - name: PRODUCT_CATALOG_SERVICE_ADDR
          value: "productcatalogservice:3550"
        - name: CURRENCY_SERVICE_ADDR
          value: "currencyservice:7000"
        - name: CART_SERVICE_ADDR
          value: "cartservice:7070"
        - name: RECOMMENDATION_SERVICE_ADDR
          value: "recommendationservice:8080"
        - name: SHIPPING_SERVICE_ADDR
          value: "shippingservice:50051"
        - name: CHECKOUT_SERVICE_ADDR
          value: "checkoutservice:5050"
        - name: AD_SERVICE_ADDR
          value: "adservice:9555"
        resources:
          requests:
            cpu: 50m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: frontend
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: frontend
  ports:
  - name: http
    port: 80
    targetPort: 8080
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: paymentservice
  labels:
    app: online-boutique
    service: paymentservice
spec:
  selector:
    matchLabels:
      app: paymentservice
  template:
    metadata:
      labels:
        app: paymentservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/paymentservice:v0.6.0
        ports:
        - containerPort: 50051
        env:
        - name: PORT
          value: "50051"
        resources:
          requests:
            cpu: 50m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: paymentservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: paymentservice
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: productcatalogservice
  labels:
    app: online-boutique
    service: productcatalogservice
spec:
  selector:
    matchLabels:
      app: productcatalogservice
  template:
    metadata:
      labels:
        app: productcatalogservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/productcatalogservice:v0.6.0
        ports:
        - containerPort: 3550
        env:
        - name: PORT
          value: "3550"
        resources:
          requests:
            cpu: 50m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: productcatalogservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: productcatalogservice
  ports:
  - name: grpc
    port: 3550
    targetPort: 3550
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: cartservice
  labels:
    app: online-boutique
    service: cartservice
spec:
  selector:
    matchLabels:
      app: cartservice
  template:
    metadata:
      labels:
        app: cartservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/cartservice:v0.6.0
        ports:
        - containerPort: 7070
        env:
        - name: REDIS_ADDR
          value: "redis-cart:6379"
        resources:
          requests:
            cpu: 100m
            memory: 32Mi
          limits:
            cpu: 150m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: cartservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: cartservice
  ports:
  - name: grpc
    port: 7070
    targetPort: 7070
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: currencyservice
  labels:
    app: online-boutique
    service: currencyservice
spec:
  selector:
    matchLabels:
      app: currencyservice
  template:
    metadata:
      labels:
        app: currencyservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/currencyservice:v0.6.0
        ports:
        - containerPort: 7000
        env:
        - name: PORT
          value: "7000"
        resources:
          requests:
            cpu: 50m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: currencyservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: currencyservice
  ports:
  - name: grpc
    port: 7000
    targetPort: 7000
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: shippingservice
  labels:
    app: online-boutique
    service: shippingservice
spec:
  selector:
    matchLabels:
      app: shippingservice
  template:
    metadata:
      labels:
        app: shippingservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/shippingservice:v0.6.0
        ports:
        - containerPort: 50051
        env:
        - name: PORT
          value: "50051"
        resources:
          requests:
            cpu: 50m
            memory: 32Mi
          limits:
            cpu: 100m
            memory: 64Mi
---
apiVersion: v1
kind: Service
metadata:
  name: shippingservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: shippingservice
  ports:
  - name: grpc
    port: 50051
    targetPort: 50051
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: redis-cart
  labels:
    app: online-boutique
    service: redis-cart
spec:
  selector:
    matchLabels:
      app: redis-cart
  template:
    metadata:
      labels:
        app: redis-cart
    spec:
      containers:
      - name: redis
        image: redis:alpine
        ports:
        - containerPort: 6379
        resources:
          requests:
            cpu: 35m
            memory: 100Mi
          limits:
            cpu: 70m
            memory: 128Mi
        volumeMounts:
        - name: redis-data
          mountPath: /data
      volumes:
      - name: redis-data
        emptyDir: {}
---
apiVersion: v1
kind: Service
metadata:
  name: redis-cart
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: redis-cart
  ports:
  - name: redis
    port: 6379
    targetPort: 6379
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: adservice
  labels:
    app: online-boutique
    service: adservice
spec:
  selector:
    matchLabels:
      app: adservice
  template:
    metadata:
      labels:
        app: adservice
    spec:
      containers:
      - name: server
        image: gcr.io/google-samples/microservices-demo/adservice:v0.6.0
        ports:
        - containerPort: 9555
        env:
        - name: PORT
          value: "9555"
        resources:
          requests:
            cpu: 100m
            memory: 90Mi
          limits:
            cpu: 150m
            memory: 150Mi
---
apiVersion: v1
kind: Service
metadata:
  name: adservice
  labels:
    app: online-boutique
spec:
  type: ClusterIP
  selector:
    app: adservice
  ports:
  - name: grpc
    port: 9555
    targetPort: 9555
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: log-generator
  labels:
    app: online-boutique
    service: log-generator
spec:
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
        args: ["-c", "counter=1; while true; do echo \"$(date) [INFO] Log entry #$counter - Online Boutique service running smoothly\"; echo \"$(date) [DEBUG] Processing checkout request ID: $((counter * 7 % 9999))\"; echo \"$(date) [WARN] Memory usage at $((counter % 40 + 60))%\"; if [ $((counter % 10)) -eq 0 ]; then echo \"$(date) [ERROR] Simulated payment gateway timeout\"; fi; echo \"$(date) [TRACE] Cache hit ratio: $((90 + counter % 10))%\"; counter=$((counter + 1)); sleep 5; done"]
        resources:
          requests:
            cpu: 25m
            memory: 16Mi
          limits:
            cpu: 50m
            memory: 32Mi
EOF
    
    $CLI_TOOL apply -f boutique-quota-friendly.yaml
    rm -f boutique-quota-friendly.yaml
fi

echo ""
echo "⏳ Waiting for deployments to start..."
sleep 10

echo ""
echo "📈 Checking deployment status..."
$CLI_TOOL get deployments -l app=online-boutique --no-headers | while read line; do
    echo "  $line"
done

echo ""
echo "⏳ Waiting for pods to be ready (this may take a few minutes)..."
echo "   ⏰ Pulling container images and starting services..."

# Wait for pods to be ready (with timeout)
TIMEOUT=300  # 5 minutes
ELAPSED=0
while [ $ELAPSED -lt $TIMEOUT ]; do
    READY_PODS=$($CLI_TOOL get pods --no-headers 2>/dev/null | grep -E "(emailservice|checkoutservice|recommendationservice|frontend|paymentservice|productcatalogservice|cartservice|currencyservice|shippingservice|redis-cart|adservice|log-generator)" | grep -c "Running\|Completed" || echo "0")
    TOTAL_PODS=$($CLI_TOOL get pods --no-headers 2>/dev/null | grep -E "(emailservice|checkoutservice|recommendationservice|frontend|paymentservice|productcatalogservice|cartservice|currencyservice|shippingservice|redis-cart|adservice|log-generator)" | wc -l || echo "0")
    
    if [ "$READY_PODS" -gt 0 ] && [ "$READY_PODS" -eq "$TOTAL_PODS" ]; then
        echo "  ✅ All $READY_PODS pods are ready!"
        break
    fi
    
    echo "  ⏳ $READY_PODS/$TOTAL_PODS pods ready..."
    sleep 10
    ELAPSED=$((ELAPSED + 10))
done

if [ $ELAPSED -ge $TIMEOUT ]; then
    echo "  ⚠️  Timeout waiting for all pods. Continuing with current state..."
fi

echo ""
echo "🔧 Phase 3: Adding Demo Secrets & ConfigMaps"
echo "============================================="

# Add some demo secrets for LazyOC secret viewing functionality
echo "  🔒 Creating demo application secrets..."

$CLI_TOOL create secret generic payment-secrets \
    --from-literal=stripe-api-key=sk_test_51234567890abcdefghijklmnopqrstuvwxyz \
    --from-literal=paypal-client-secret=paypal-secret-key-for-payments \
    --from-literal=square-access-token=sandbox-sq0atb-1234567890abcdefghij \
    > /dev/null 2>&1 || echo "  ⚠️  Payment secrets already exist"

$CLI_TOOL create secret generic database-credentials \
    --from-literal=redis-password=boutique-redis-password-2024 \
    --from-literal=postgres-password=boutique-db-password-2024 \
    --from-literal=mongodb-password=boutique-mongo-password-2024 \
    > /dev/null 2>&1 || echo "  ⚠️  Database credentials already exist"

$CLI_TOOL create secret generic external-apis \
    --from-literal=google-analytics-key=GA-1234567-1 \
    --from-literal=sendgrid-api-key=SG.1234567890abcdefghijklmnopqrstuvwxyz \
    --from-literal=aws-access-key=AKIAIOSFODNN7EXAMPLE \
    --from-literal=aws-secret-key=wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY \
    > /dev/null 2>&1 || echo "  ⚠️  External API secrets already exist"

# Add demo configmaps
echo "  📋 Creating demo configuration maps..."

$CLI_TOOL create configmap app-settings \
    --from-literal=environment=production \
    --from-literal=debug-mode=false \
    --from-literal=log-level=info \
    --from-literal=max-connections=1000 \
    --from-literal=session-timeout=3600 \
    --from-literal=cache-ttl=7200 \
    > /dev/null 2>&1 || echo "  ⚠️  App settings ConfigMap already exists"

$CLI_TOOL create configmap feature-flags \
    --from-literal=enable-recommendations=true \
    --from-literal=enable-ads=true \
    --from-literal=enable-profiler=false \
    --from-literal=enable-tracing=true \
    --from-literal=currency-service-enabled=true \
    > /dev/null 2>&1 || echo "  ⚠️  Feature flags ConfigMap already exists"

# Label the new resources
$CLI_TOOL label secrets payment-secrets database-credentials external-apis demo=lazyoc-production > /dev/null 2>&1 || true
$CLI_TOOL label configmaps app-settings feature-flags demo=lazyoc-production > /dev/null 2>&1 || true

# Only create OpenShift-specific resources if this is an OpenShift cluster
if [ "$CLI_TOOL" = "oc" ] && oc api-resources | grep -q "routes.*route.openshift.io"; then
    echo ""
    echo "🌐 Phase 4: Creating OpenShift Routes"
    echo "===================================="
    
    echo "  🌍 Creating route for frontend service..."
    cat <<EOF | oc apply -f - > /dev/null 2>&1 || echo "  ⚠️  Route creation failed"
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: online-boutique-frontend
  labels:
    demo: lazyoc-production
spec:
  to:
    kind: Service
    name: frontend-external
  port:
    targetPort: http
  tls:
    termination: edge
    insecureEdgeTerminationPolicy: Redirect
EOF

    OPENSHIFT_DEPLOYED=true
else
    echo ""
    echo "ℹ️  Skipping OpenShift-specific resources (not an OpenShift cluster)"
    OPENSHIFT_DEPLOYED=false
fi

echo ""
echo "📊 Phase 5: Final Status & Resource Summary"
echo "==========================================="

echo ""
echo "🏃 Pods Status:"
echo "==============="
$CLI_TOOL get pods --no-headers | grep -E "(emailservice|checkoutservice|recommendationservice|frontend|paymentservice|productcatalogservice|cartservice|currencyservice|shippingservice|redis-cart|adservice|log-generator)" | while read line; do
    echo "  $line"
done

echo ""
echo "🚀 Deployments Status:"
echo "======================"
$CLI_TOOL get deployments --no-headers | grep -E "(emailservice|checkoutservice|recommendationservice|frontend|paymentservice|productcatalogservice|cartservice|currencyservice|shippingservice|redis-cart|adservice|log-generator)" | while read line; do
    echo "  $line"
done

echo ""
echo "🌐 Services:"
echo "============"
$CLI_TOOL get services --no-headers | grep -E "(emailservice|checkoutservice|recommendationservice|frontend|paymentservice|productcatalogservice|cartservice|currencyservice|shippingservice|redis-cart|adservice)" | while read line; do
    echo "  $line"
done

echo ""
echo "📋 ConfigMaps:"
echo "=============="
$CLI_TOOL get configmaps -l demo=lazyoc-production --no-headers | while read line; do
    echo "  $line"
done

echo ""
echo "🔐 Secrets:"
echo "==========="
$CLI_TOOL get secrets -l demo=lazyoc-production --no-headers | while read line; do
    echo "  $line"
done

if [ "$OPENSHIFT_DEPLOYED" = true ]; then
    echo ""
    echo "🌍 Routes:"
    echo "=========="
    oc get routes -l demo=lazyoc-production --no-headers | while read line; do
        echo "  $line"
    done
    
    echo ""
    echo "🌐 Access the application:"
    ROUTE_URL=$(oc get route online-boutique-frontend -o jsonpath='{.spec.host}' 2>/dev/null || echo "Route not found")
    if [ "$ROUTE_URL" != "Route not found" ]; then
        echo "  🌍 https://$ROUTE_URL"
    fi
fi

# Cleanup temp directory
cd - > /dev/null
rm -rf "$TEMP_DIR"

echo ""
echo "✅ Production LazyOC Demo Environment Ready!"
echo "============================================"
echo ""
echo "🎯 Google Cloud Online Boutique Application Deployed:"
echo ""
echo "📱 Microservices Architecture (12 services):"
echo "  • 🏪 Frontend - Web UI (React/TypeScript)"  
echo "  • 🛒 Cart Service - Shopping cart management"
echo "  • 💳 Checkout Service - Payment processing" 
echo "  • 🎯 Recommendation Service - Product recommendations"
echo "  • 📦 Product Catalog Service - Product information"
echo "  • 💰 Currency Service - Multi-currency support"
echo "  • 💸 Payment Service - Payment processing"
echo "  • 📧 Email Service - Order confirmations"
echo "  • 🚚 Shipping Service - Shipping calculations"
echo "  • 📊 Ad Service - Advertisement management"
echo "  • 🗂️  Redis Cache - Session and cart storage"
echo "  • 📋 Log Generator - Continuous logging for testing"
echo ""
echo "🔧 LazyOC Resource Types Available:"
echo "  1. 🏃 Pods - 12+ pods across all microservices"
echo "  2. 🌐 Services - 11+ interconnected services"  
echo "  3. 🚀 Deployments - 12 production deployments"
echo "  4. 📋 ConfigMaps - Feature flags & app settings"
echo "  5. 🔐 Secrets - Payment keys & database credentials"
if [ "$OPENSHIFT_DEPLOYED" = true ]; then
echo "  6. 🌍 Routes - External access routes"
fi
echo ""
echo "🔍 Special Features to Test:"
echo ""
echo "🔥 Service Log Aggregation (Key Feature):"
echo "  • Navigate to Services tab → Select 'frontend' → Press 'L'"
echo "  • Shows logs from multiple frontend pods aggregated"
echo "  • Try with 'cartservice', 'checkoutservice', 'recommendationservice'"
echo "  • Real production microservice logs with business logic"
echo ""
echo "🔐 Secret Management:"
echo "  • Navigate to Secrets tab → Select 'payment-secrets' → Press 'Enter'"  
echo "  • Use 'j/k' to navigate, 'm' to toggle masking, 'c' to copy"
echo "  • Test with 'database-credentials', 'external-apis'"
echo "  • Real-world credential structures and key formats"
echo ""
echo "📊 Rich Production Logs:"
echo "  • gRPC communication between services"
echo "  • Request tracing and performance metrics"
echo "  • Business logic: cart operations, payments, recommendations"
echo "  • Error handling and circuit breaker patterns"
echo "  • Health checks and readiness probes"
echo ""
echo "🧪 Comprehensive Testing Flow:"
echo "  1. 🔧 Build: go build -o ./bin/lazyoc ./cmd/lazyoc"
echo "  2. 🚀 Run: ./bin/lazyoc"
echo "  3. 📱 Explore all resource tabs with real production data"
echo "  4. 🔍 Test service log aggregation on multi-pod services"
echo "  5. 🔐 Test secret viewing with realistic credentials"
echo "  6. ⚙️  Test auto-refresh with active application traffic"
echo "  7. 🔄 Navigate between services to see microservice interactions"
echo ""
echo "💡 Application Features:"
echo "  • Complete e-commerce workflow (browse → cart → checkout)"
echo "  • Multi-language support and currency conversion"
echo "  • Recommendation engine with product suggestions"
echo "  • Real payment processing simulation"
echo "  • Distributed tracing and observability"
echo ""
echo "🧹 Cleanup: $CLI_TOOL delete -f https://raw.githubusercontent.com/GoogleCloudPlatform/microservices-demo/v0.6.0/release/kubernetes-manifests.yaml"
echo "   Additional: $CLI_TOOL delete secrets,configmaps -l demo=lazyoc-production"
echo ""