#!/usr/bin/env bash

# Simple release script for k8s-controller
# Uses git-cliff to generate changelog and determine next version

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" &> /dev/null && pwd)
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
cd "$PROJECT_ROOT"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Check if git-cliff is installed
if ! command -v git-cliff &> /dev/null; then
    echo "❌ git-cliff is not installed. Please install it first:"
    echo "   cargo install git-cliff"
    echo "   or"
    echo "   brew install git-cliff"
    exit 1
fi

log_info "Checking repository status..."

# Check if any tags exist
if ! git describe --tags --abbrev=0 &>/dev/null; then
    log_warning "No tags found in repository."
    echo "This appears to be the first release."
    echo ""
    log_info "Recommended next steps:"
    echo "1. Generate changelog and create first tag:"
    echo "   git cliff --tag v0.1.0 --output CHANGELOG.md"
    echo "   git add CHANGELOG.md"
    echo "   git commit -m 'chore(release): prepare for v0.1.0'"
    echo "   git tag -a v0.1.0 -m 'Release v0.1.0'"
    echo "   git push origin main && git push origin v0.1.0"
    exit 0
fi

# Get current version
current_version=$(git describe --tags --abbrev=0)
log_info "Current version: $current_version"

# Use git-cliff to determine next version and generate changelog
log_info "Using git-cliff to determine next version..."

# Generate new changelog with bumped version
if git cliff --bump --output CHANGELOG.md; then
    log_success "CHANGELOG.md updated successfully!"

    # Extract the new version from the generated changelog
    next_version=$(head -10 CHANGELOG.md | grep -oE '\[([0-9]+\.[0-9]+\.[0-9]+)\]' | head -1 | tr -d '[]')

    if [ -n "$next_version" ]; then
        next_version="v${next_version}"
        log_info "Next version will be: $next_version"

        echo ""
        log_info "📋 Changelog preview:"
        echo "----------------------------------------"
        head -20 CHANGELOG.md
        echo "----------------------------------------"
        echo ""

        log_success "✅ Ready for release!"
        echo ""
        log_info "🚀 Recommended next steps:"
        echo "1. Review the generated CHANGELOG.md"
        echo "2. Commit and create the release:"
        echo "   git add CHANGELOG.md"
        echo "   git commit -m 'chore(release): prepare for $next_version'"
        echo "   git tag -a $next_version -m 'Release $next_version'"
        echo "   git push origin main && git push origin $next_version"
        echo ""
        echo "Or run these commands all at once:"
        echo "   git add CHANGELOG.md && \\"
        echo "   git commit -m 'chore(release): prepare for $next_version' && \\"
        echo "   git tag -a $next_version -m 'Release $next_version' && \\"
        echo "   git push origin main && git push origin $next_version"
    else
        log_warning "Could not determine next version from changelog"
        log_info "Please check CHANGELOG.md manually"
    fi
else
    log_warning "git-cliff could not generate changelog"
    echo "This usually means no conventional commits found since last release."
    echo ""
    log_info "You can force a patch release with:"
    echo "   git cliff --tag v$(echo "$current_version" | sed 's/v//' | awk -F. '{print $1"."$2"."$3+1}') --output CHANGELOG.md"
fi
