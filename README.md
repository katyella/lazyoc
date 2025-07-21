# LazyOC 🚀

[![CI Status](https://github.com/katyella/lazyoc/workflows/CI/badge.svg)](https://github.com/katyella/lazyoc/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/katyella/lazyoc)](https://goreportcard.com/report/github.com/katyella/lazyoc)
[![codecov](https://codecov.io/gh/katyella/lazyoc/branch/main/graph/badge.svg)](https://codecov.io/gh/katyella/lazyoc)
[![Release](https://img.shields.io/github/release/katyella/lazyoc.svg)](https://github.com/katyella/lazyoc/releases/latest)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A lazy terminal UI for OpenShift and Kubernetes clusters. LazyOC provides an intuitive, vim-like interface for managing cluster resources without the complexity of memorizing kubectl/oc commands.

## 🌟 Features

### Core Functionality
- **Terminal UI**: Clean, responsive interface built with Bubble Tea
- **Multi-cluster Support**: Manage multiple OpenShift/Kubernetes clusters simultaneously
- **Real-time Updates**: Live resource monitoring with automatic refresh
- **Vim-like Navigation**: Familiar keybindings for efficient cluster management

### Resource Management
- **Resource Listing**: View pods, services, deployments, and more
- **Resource Operations**: Describe, delete, restart, and scale resources
- **Log Streaming**: Real-time container logs with filtering
- **Shell Access**: Direct container shell access via exec

### OpenShift-Specific Features
- **BuildConfigs**: Monitor and trigger builds
- **ImageStreams**: Manage container images and tags
- **Routes**: Configure application routing
- **Operators**: Manage OpenShift operators and subscriptions

### Developer Workflow
- **Port Forwarding**: Automatic tunnel management
- **File Transfer**: Bidirectional file sync with containers
- **Resource Editing**: YAML/JSON editing with validation
- **Hot Reload**: Apply configuration changes without downtime

## 🚀 Installation

### macOS (Homebrew)
```bash
brew install mattwojtowicz/tap/lazyoc
```

### Linux/macOS (Manual)
```bash
# Download the latest release
curl -L https://github.com/katyella/lazyoc/releases/latest/download/lazyoc_Linux_x86_64.tar.gz | tar xz
sudo mv lazyoc /usr/local/bin/

# Or for macOS
curl -L https://github.com/katyella/lazyoc/releases/latest/download/lazyoc_Darwin_x86_64.tar.gz | tar xz
sudo mv lazyoc /usr/local/bin/
```

### Windows (Scoop)
```powershell
scoop bucket add mattwojtowicz https://github.com/mattwojtowicz/scoop-bucket
scoop install lazyoc
```

### From Source
```bash
go install github.com/katyella/lazyoc/cmd/lazyoc@latest
```

## 🎯 Quick Start

1. **Launch LazyOC**: `lazyoc`
2. **Connect to cluster**: LazyOC will automatically detect your current kubeconfig context
3. **Navigate resources**: Use arrow keys or vim navigation (hjkl)
4. **Access help**: Press `?` for keyboard shortcuts

### Configuration

LazyOC uses your existing kubeconfig file by default. You can specify a different config:

```bash
lazyoc --kubeconfig=/path/to/config
```

## 🔐 Authentication & Kubeconfig

### Understanding Kubeconfig

The kubeconfig file is a YAML configuration that stores:
- **Clusters**: API server endpoints and certificate authorities
- **Users**: Authentication credentials (tokens, certificates)
- **Contexts**: Combinations of cluster + user + namespace
- **Current Context**: The active cluster/user/namespace

Default location: `~/.kube/config`

### Working with `oc login`

When you authenticate to an OpenShift cluster using `oc login`, it automatically updates your kubeconfig:

```bash
# Login to OpenShift
oc login https://api.cluster.example.com:6443 --token=sha256~ABC123...

# This updates ~/.kube/config with:
# - Cluster endpoint and CA certificate
# - Your authentication token
# - A new context for this cluster

# LazyOC automatically uses this configuration
lazyoc
```

### Relationship between `oc login` and LazyOC

```
┌──────────┐     writes to      ┌─────────────────┐     reads from    ┌─────────┐
│ oc login │ ─────────────────> │ ~/.kube/config  │ <───────────────── │ LazyOC  │
└──────────┘                    └─────────────────┘                    └─────────┘
                                         ↑
                                    also used by
                                         ↑
                                ┌─────────────────┐
                                │ kubectl/oc CLI  │
                                └─────────────────┘
```

### Multiple Clusters

Manage multiple clusters easily:

```bash
# Login to different clusters
oc login https://dev.openshift.com --token=dev-token
oc login https://prod.openshift.com --token=prod-token

# Switch active context
oc config use-context dev-cluster

# LazyOC uses the current context
lazyoc

# Or explicitly specify a different kubeconfig
lazyoc --kubeconfig=/path/to/other/config
```

### Authentication Methods

LazyOC supports all Kubernetes/OpenShift authentication methods:
- **OAuth Tokens**: From `oc login` (most common for OpenShift)
- **Client Certificates**: X.509 certificates
- **Service Account Tokens**: For automation
- **Basic Auth**: Username/password (if enabled)
- **OIDC Tokens**: From external identity providers

### Troubleshooting Authentication

If LazyOC cannot connect:

```bash
# Check current context
kubectl config current-context

# List all contexts
kubectl config get-contexts

# Verify cluster connectivity
kubectl cluster-info

# Re-authenticate if needed
oc login https://your-cluster.com
```

## ⚡ Performance Targets

LazyOC is designed to be lightweight and efficient:
- **Memory**: <100MB baseline usage
- **CPU**: <5% average, <10% peak
- **Startup**: <2 seconds
- **API Latency**: <500ms for queries

## 🛠 Development

### Prerequisites
- Go 1.21 or higher
- Make
- golangci-lint (for development)

### Building from Source
```bash
git clone https://github.com/katyella/lazyoc
cd lazyoc
make build
```

### Development Commands
```bash
make dev          # Run in development mode
make test         # Run tests
make lint         # Run linter
make build        # Build binary
make install      # Install to GOPATH/bin
```

### Project Structure
```
├── cmd/lazyoc/          # Application entrypoint
├── internal/            # Private application code
├── pkg/                 # Public libraries
├── api/                 # API definitions
├── configs/             # Configuration files
└── scripts/             # Build and utility scripts
```

## 🤝 Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

### Development Setup
1. Fork the repository
2. Clone your fork: `git clone https://github.com/yourusername/lazyoc`
3. Create a feature branch: `git checkout -b feature/amazing-feature`
4. Make your changes and add tests
5. Run tests: `make test`
6. Run linting: `make lint`
7. Commit your changes: `git commit -m 'Add amazing feature'`
8. Push to your fork: `git push origin feature/amazing-feature`
9. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - Terminal UI framework
- [Kubernetes](https://kubernetes.io/) - Container orchestration
- [OpenShift](https://www.redhat.com/en/technologies/cloud-computing/openshift) - Enterprise Kubernetes platform

## 🔗 Links

- [Documentation](https://github.com/katyella/lazyoc/wiki)
- [Issue Tracker](https://github.com/katyella/lazyoc/issues)
- [Discussions](https://github.com/katyella/lazyoc/discussions)
- [Releases](https://github.com/katyella/lazyoc/releases)