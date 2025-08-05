# LazyOC Demo Environment Scripts

This directory contains setup scripts to create comprehensive test environments for demonstrating LazyOC's capabilities.

## 🚀 Recommended: Production Demo

**`./setup-production-demo.sh`** - Google Cloud Online Boutique (11 microservices)

- **Best for**: Complete LazyOC feature demonstration
- **What it deploys**: Production-ready e-commerce application with 11 microservices
- **Technologies**: gRPC, React, Go, Java, Python, C#, Node.js
- **Resource types**: All 8 LazyOC tabs (Pods, Services, Deployments, ConfigMaps, Secrets, BuildConfigs*, ImageStreams*, Routes*)
- **Special features**: 
  - Service log aggregation from multiple pods
  - Real microservice communication logs
  - Production-ready secret structures
  - Load balancing across multiple replicas
- **Cleanup**: `./teardown-production-demo.sh`

*\*OpenShift-specific resources only created on OpenShift clusters*

## 📦 Alternative Options

### Simple Test Apps (Legacy)
**`./setup-test-apps.sh`** - Basic test applications

- **Best for**: Quick testing of basic LazyOC functionality  
- **What it deploys**: Nginx, Hello World, Redis, Apache, Log Generator + 7 test secrets
- **Resource types**: Pods, Services, Deployments, Secrets, ConfigMaps
- **Cleanup**: `./teardown-test-apps.sh`

### Kubernetes-Only Secrets
**`./setup-k8s-test-secrets.sh`** - Standalone secret testing

- **Best for**: Testing secret viewing functionality only
- **What it creates**: 8 comprehensive test secrets with various data types
- **Compatible with**: Any Kubernetes cluster (uses `kubectl`)
- **Cleanup**: Manual deletion of created secrets

## 🧪 Usage Instructions

### Quick Start (Recommended)
```bash
# Deploy production demo
./scripts/setup-production-demo.sh

# Run LazyOC
go build -o ./bin/lazyoc ./cmd/lazyoc && ./bin/lazyoc

# Clean up when done
./scripts/teardown-production-demo.sh
```

### Testing Workflow

1. **Deploy environment** using one of the setup scripts
2. **Build LazyOC**: `go build -o ./bin/lazyoc ./cmd/lazyoc`
3. **Run LazyOC**: `./bin/lazyoc`
4. **Test key features**:
   - Navigate through all resource tabs
   - Service log aggregation: Services tab → Select service → Press 'L'
   - Secret viewing: Secrets tab → Select secret → Press 'Enter'
   - Project switching: Press 'Ctrl+P'
   - Auto-refresh and manual refresh ('r' key)

## 🎯 Feature Testing Guide

### Service Log Aggregation
- **Best services to test**: `frontend`, `cartservice`, `checkoutservice`, `recommendationservice`
- **What to look for**: Logs from multiple pods aggregated with `[pod-name/container]` prefixes
- **Navigation**: Press 'L' on Services tab, 'j/k' to scroll logs

### Secret Management  
- **Best secrets to test**: `payment-secrets`, `database-credentials`, `external-apis`
- **What to test**: Masking ('m' key), copying ('c' for value, 'C' for JSON), navigation ('j/k')
- **Security features**: Masked by default, clipboard integration

### Resource Monitoring
- **Auto-refresh**: Resources update every 30 seconds automatically
- **Manual refresh**: Press 'r' key to force refresh
- **Real-time logs**: Log streams update continuously when viewing pod logs

## 🔧 Requirements

- **Kubernetes/OpenShift cluster** with kubectl/oc access
- **Internet connectivity** for downloading container images
- **Sufficient resources**: ~2-4 GB RAM, ~2 CPU cores for production demo
- **Optional**: Git for cloning repositories (fallback downloads available)

## 🧹 Cleanup

Each setup script has a corresponding teardown script:
- `setup-production-demo.sh` → `teardown-production-demo.sh`
- `setup-test-apps.sh` → `teardown-test-apps.sh`
- `setup-k8s-test-secrets.sh` → Manual cleanup (commands provided in script output)

## 💡 Tips

- Use **production demo** for the most comprehensive testing experience
- **OpenShift clusters** get additional resources (Routes, BuildConfigs, ImageStreams)
- **Service log aggregation** works best with multi-replica services
- **Secret viewing** supports various data types (JSON, certificates, credentials)
- Scripts are **idempotent** - safe to run multiple times

## 📋 Available Scripts Summary

| Script | Purpose | Resources | Cleanup |
|--------|---------|-----------|---------|
| `setup-production-demo.sh` | 🏆 **Recommended** - Full demo | 11 microservices, all resource types | `teardown-production-demo.sh` |
| `setup-test-apps.sh` | Simple legacy testing | 5 basic apps + secrets | `teardown-test-apps.sh` |
| `setup-k8s-test-secrets.sh` | Secret testing only | 8 test secrets | Manual deletion |

## 🚀 Quick Commands

```bash
# Production demo (recommended)
./scripts/setup-production-demo.sh
go build -o ./bin/lazyoc ./cmd/lazyoc && ./bin/lazyoc
./scripts/teardown-production-demo.sh

# Legacy test apps
./scripts/setup-test-apps.sh
go build -o ./bin/lazyoc ./cmd/lazyoc && ./bin/lazyoc  
./scripts/teardown-test-apps.sh

# Secrets only
./scripts/setup-k8s-test-secrets.sh
go build -o ./bin/lazyoc ./cmd/lazyoc && ./bin/lazyoc
# Manual cleanup as shown in script output
```