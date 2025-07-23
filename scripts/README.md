# LazyOC Test Scripts

This directory contains scripts to help you quickly set up and tear down test applications in OpenShift for testing LazyOC functionality.

## Scripts

### `setup-test-apps.sh`
Sets up multiple test applications in your OpenShift cluster for testing LazyOC's pod listing and project management features.

**What it deploys:**
- **nginx-test** (2 replicas) - Web server for testing
- **hello-world** (3 replicas) - Simple hello world app
- **redis-test** (1 replica) - Redis database
- **httpd-test** (1 replica) - Apache HTTP server

**Usage:**
```bash
./scripts/setup-test-apps.sh
```

**Prerequisites:**
- OpenShift CLI (`oc`) installed
- Logged into OpenShift cluster (`oc login`)
- Access to an OpenShift project/namespace

### `teardown-test-apps.sh`
Removes all test applications and associated resources deployed by the setup script.

**Usage:**
```bash
./scripts/teardown-test-apps.sh
```

**What it cleans up:**
- All test deployments
- Associated pods, services, and routes
- Replica sets and other generated resources

## Testing Workflow

### 1. Setup Test Environment
```bash
# Make sure you're logged into OpenShift
oc login <your-openshift-cluster>

# Deploy test applications
./scripts/setup-test-apps.sh
```

### 2. Test LazyOC
```bash
# Build LazyOC
go build ./cmd/lazyoc

# Run LazyOC against your test environment
./lazyoc
```

### 3. Verify LazyOC Features
- ✅ **OpenShift Detection**: Should detect cluster as OpenShift (not Kubernetes)
- 🎯 **Project Context**: Header shows OpenShift project with correct icon
- 📦 **Pod Listing**: Shows all test pods with status indicators
- ⏱️ **Auto-refresh**: Pods refresh every 30 seconds
- 🔄 **Manual Refresh**: Press 'r' to refresh pod list
- ⌨️ **Navigation**: Arrow keys to select pods
- 🎮 **Project Switching**: Ctrl+P opens project modal

### 4. Clean Up
```bash
# Remove all test applications
./scripts/teardown-test-apps.sh
```

## Expected Pod Status Display

With the test apps running, you should see something like:
```
📦 Pods in your-project-name

NAME                        STATUS    READY   AGE
✅ nginx-test-1-abcde       Running   1/1     2m
✅ nginx-test-2-fghij       Running   1/1     2m  
✅ hello-world-1-klmno      Running   1/1     1m
✅ hello-world-2-pqrst      Running   1/1     1m
✅ hello-world-3-uvwxy      Running   1/1     1m
✅ redis-test-1-zabcd       Running   1/1     30s
✅ httpd-test-1-efghi       Running   1/1     30s
```

## Troubleshooting

### Script Permission Issues
```bash
chmod +x scripts/*.sh
```

### OpenShift Login Issues
```bash
# Check if logged in
oc whoami

# Check current project
oc project
```

### Resource Quota Issues
If deployments fail due to resource limits in OpenShift Sandbox:
- The sandbox has limits: 3 cores, 14GB RAM, 40GB storage
- Reduce replica counts in the setup script if needed
- Some pods may stay in Pending state due to resource constraints

### Application Already Exists
If you see "already exists" warnings:
- Run teardown script first: `./scripts/teardown-test-apps.sh`
- Then run setup script again: `./scripts/setup-test-apps.sh`

## OpenShift Sandbox Specific Notes

When using Red Hat OpenShift Developer Sandbox:
- **Resource Limits**: 3 cores, 14GB RAM, 40GB storage
- **Pod Lifetime**: Pods auto-delete after 12 hours of runtime
- **Perfect for Testing**: Provides real OpenShift project APIs
- **Web Console**: Available for visual verification at console-openshift-console.apps-crc.testing

These scripts are designed to work well within the Developer Sandbox resource constraints while providing enough variety to thoroughly test LazyOC's OpenShift integration.