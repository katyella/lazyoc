#!/bin/bash

echo "🚀 Setting up OpenShift resources for LazyOC testing..."

# Check if we're logged in
if ! oc whoami &>/dev/null; then
    echo "❌ Not logged into OpenShift. Please run 'oc login' first."
    exit 1
fi

echo "📍 Current project: $(oc project -q)"

# Create a simple BuildConfig
echo "🔨 Creating BuildConfig..."
cat <<EOF | oc apply -f -
apiVersion: build.openshift.io/v1
kind: BuildConfig
metadata:
  name: sample-app-build
  labels:
    app: sample-app
spec:
  source:
    type: Git
    git:
      uri: https://github.com/openshift/ruby-hello-world
  strategy:
    type: Source
    sourceStrategy:
      from:
        kind: ImageStreamTag
        name: ruby:3.0
        namespace: openshift
  output:
    to:
      kind: ImageStreamTag
      name: sample-app:latest
EOF

# Create another BuildConfig
echo "🔨 Creating another BuildConfig..."
cat <<EOF | oc apply -f -
apiVersion: build.openshift.io/v1
kind: BuildConfig
metadata:
  name: nginx-build
  labels:
    app: nginx-app
spec:
  source:
    type: Git
    git:
      uri: https://github.com/sclorg/nginx-container
      ref: master
    contextDir: examples/nginx-test-app
  strategy:
    type: Source
    sourceStrategy:
      from:
        kind: ImageStreamTag
        name: nginx:1.20
        namespace: openshift
  output:
    to:
      kind: ImageStreamTag
      name: nginx-app:latest
EOF

# Create a Route
echo "🛣️ Creating Route..."
cat <<EOF | oc apply -f -
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: sample-app-route
  labels:
    app: sample-app
spec:
  host: sample-app-$(oc project -q).apps.rm3.7wse.p1.openshiftapps.com
  to:
    kind: Service
    name: sample-app-service
  port:
    targetPort: 8080-tcp
  tls:
    termination: edge
    insecureEdgeTerminationPolicy: Redirect
EOF

# Create another Route (HTTP only)
echo "🛣️ Creating HTTP Route..."
cat <<EOF | oc apply -f -
apiVersion: route.openshift.io/v1
kind: Route
metadata:
  name: nginx-route
  labels:
    app: nginx-app
spec:
  host: nginx-$(oc project -q).apps.rm3.7wse.p1.openshiftapps.com
  to:
    kind: Service
    name: nginx-service
  port:
    targetPort: 8080-tcp
EOF

echo "✅ OpenShift resources created successfully!"
echo ""
echo "📋 Summary:"
echo "  🔨 BuildConfigs: sample-app-build, nginx-build"
echo "  🛣️ Routes: sample-app-route, nginx-route"
echo "  🖼️ ImageStreams: $(oc get imagestreams -o name | wc -l) existing"
echo ""
echo "🎯 Now run LazyOC and navigate to the OpenShift tabs to see the resources!"
echo "   ./bin/lazyoc"