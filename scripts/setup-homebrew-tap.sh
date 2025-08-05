#!/bin/bash

# Script to set up Homebrew tap repository for LazyOC
set -e

REPO_OWNER="katyella"
TAP_REPO="homebrew-tap"
CURRENT_DIR=$(pwd)

echo "🍺 Setting up Homebrew tap repository..."

# Check if gh CLI is available
if ! command -v gh &> /dev/null; then
    echo "❌ GitHub CLI (gh) is required. Install with: brew install gh"
    exit 1
fi

# Check if user is authenticated
if ! gh auth status &> /dev/null; then
    echo "❌ Please authenticate with GitHub CLI: gh auth login"
    exit 1
fi

# Create the tap repository if it doesn't exist
echo "📦 Creating tap repository..."
if gh repo view "$REPO_OWNER/$TAP_REPO" &> /dev/null; then
    echo "✅ Repository $REPO_OWNER/$TAP_REPO already exists"
else
    gh repo create "$REPO_OWNER/$TAP_REPO" --public --description "Homebrew tap for LazyOC"
    echo "✅ Created repository $REPO_OWNER/$TAP_REPO"
fi

# Clone and set up the tap repository
echo "📂 Setting up tap structure..."
cd /tmp
rm -rf "$TAP_REPO" 2>/dev/null || true
gh repo clone "$REPO_OWNER/$TAP_REPO"
cd "$TAP_REPO"

# Create Formula directory
mkdir -p Formula

# Create README.md
cat > README.md << 'EOF'
# LazyOC Homebrew Tap

This is the Homebrew tap for [LazyOC](https://github.com/katyella/lazyoc), a lazy terminal UI for OpenShift/Kubernetes clusters.

## Installation

```bash
brew tap katyella/tap
brew install lazyoc
```

## Updating

```bash
brew update
brew upgrade lazyoc
```

## Manual Installation

If you prefer not to use the tap:

```bash
# Download latest release
curl -L https://github.com/katyella/lazyoc/releases/latest/download/lazyoc_$(uname -s)_$(uname -m).tar.gz | tar xz
sudo mv lazyoc /usr/local/bin/
```
EOF

# Commit and push if there are changes
if [ -n "$(git status --porcelain)" ]; then
    git add .
    git commit -m "Initial tap setup"
    git push origin main
    echo "✅ Pushed initial tap setup"
else
    echo "✅ Tap repository already up to date"
fi

cd "$CURRENT_DIR"

echo ""
echo "🎉 Homebrew tap setup complete!"
echo ""
echo "Next steps:"
echo "1. Create a new release to trigger formula generation:"
echo "   git tag v0.1.1"
echo "   git push origin v0.1.1"
echo ""
echo "2. After release completes, test installation:"
echo "   brew tap $REPO_OWNER/tap"
echo "   brew install lazyoc"
echo ""
echo "3. Users can then install with:"
echo "   brew tap $REPO_OWNER/tap"
echo "   brew install lazyoc"