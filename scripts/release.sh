#!/usr/bin/env bash

# Release script for k8s-controller
# This script helps create releases with proper changelog generation

set -euo pipefail

SCRIPT_DIR=$(cd -- "$(dirname -- "$0")" &> /dev/null && pwd)
PROJECT_ROOT=$(dirname "$SCRIPT_DIR")
cd "$PROJECT_ROOT"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Git related
current_branch=$(git branch --show-current)
local_commit=$(git rev-parse HEAD)
remote_commit=$(git rev-parse origin/main)
current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if git-cliff is installed
check_git_cliff() {
    if ! command -v git-cliff &> /dev/null; then
        log_error "git-cliff is not installed. Please install it first:"
        echo "  cargo install git-cliff"
        echo "  or"
        echo "  brew install git-cliff"
        exit 1
    fi
}

# Check if we're on main branch and up to date
check_git_status() {
    if [ "$current_branch" != "main" ]; then
        log_error "You must be on the main branch to create a release"
        exit 1
    fi

    if [ -n "$(git status --porcelain)" ]; then
        log_error "Working directory is not clean. Please commit or stash your changes."
        exit 1
    fi

    log_info "Fetching latest changes..."
    git fetch origin

    if [ "$local_commit" != "$remote_commit" ]; then
        log_error "Your local main branch is not up to date with origin/main"
        log_info "Please run: git pull origin main"
        exit 1
    fi
}

# Get the next version
get_next_version() {
    echo "Current version: $current_version"
    echo ""
    echo "What type of release is this?"
    echo "1) Patch (bug fixes)          - ${current_version} -> ${current_version%.*}.$((${current_version##*.} + 1))"
    echo "2) Minor (new features)       - ${current_version} -> ${current_version%.*.*}.$((${current_version##*.*.} + 1)).0"
    echo "3) Major (breaking changes)   - ${current_version} -> v$((${current_version#v*.*.*} + 1)).0.0"
    echo "4) Custom version"
    echo ""

    read -pr "Enter your choice (1-4): " choice

    case $choice in
        1)
            # Patch version
            local version_num=${current_version#v}
            local patch=${version_num##*.}
            local base=${version_num%.*}
            next_version="v${base}.$((patch + 1))"
            ;;
        2)
            # Minor version
            local version_num=${current_version#v}
            local minor=${version_num%.*}
            minor=${minor##*.}
            local major=${version_num%%.*}
            next_version="v${major}.$((minor + 1)).0"
            ;;
        3)
            # Major version
            local version_num=${current_version#v}
            local major=${version_num%%.*}
            next_version="v$((major + 1)).0.0"
            ;;
        4)
            read -pr "Enter custom version (e.g., v1.2.3): " next_version
            if [[ ! $next_version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
                log_error "Invalid version format. Use vX.Y.Z format."
                exit 1
            fi
            ;;
        *)
            log_error "Invalid choice"
            exit 1
            ;;
    esac

    echo "$next_version"
}

# Generate changelog
generate_changelog() {
    local version=$1

    log_info "Generating changelog for version $version..."

    # Generate full changelog
    git-cliff --tag "$version" --output CHANGELOG.md

    log_success "Changelog generated successfully"

    # Show the changelog for this version
    echo ""
    log_info "Changelog for $version:"
    echo "----------------------------------------"
    git-cliff --tag "$version" --unreleased
    echo "----------------------------------------"
    echo ""
}

# Main release function
create_release() {
    local version=$1

    log_info "Creating release $version..."

    # Run tests to make sure everything is working
    log_info "Running tests..."
    if ! go test ./...; then
        log_error "Tests failed. Please fix them before creating a release."
        exit 1
    fi

    # Generate changelog
    generate_changelog "$version"

    # Ask for confirmation
    read -pr "Do you want to proceed with creating release $version? (y/N): " confirm
    if [[ $confirm != [yY] ]]; then
        log_warning "Release cancelled"
        exit 0
    fi

    # Commit changelog
    git add CHANGELOG.md
    git commit -m "chore(release): prepare for $version"

    # Create and push tag
    git tag -a "$version" -m "Release $version"

    # Push changes
    log_info "Pushing changes and tag..."
    git push origin main
    git push origin "$version"

    log_success "Release $version created successfully!"
    log_info "GitHub Actions will now build and publish the release."
    log_info "Check the progress at: https://github.com/$(git config remote.origin.url | sed 's|.*[:/]||' | sed 's|\.git||')/actions"
}

# Main script
main() {
    log_info "Starting release process for k8s-controller..."

    check_git_cliff
    check_git_status

    version=$(get_next_version)

    create_release "$version"
}

# Show help
if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    echo "Release script for k8s-controller"
    echo ""
    echo "Usage: $0"
    echo ""
    echo "This script will:"
    echo "  1. Check that you're on main branch and up to date"
    echo "  2. Run tests to ensure code quality"
    echo "  3. Generate changelog using git-cliff"
    echo "  4. Create a new git tag"
    echo "  5. Push changes and tag to trigger GitHub Actions release"
    echo ""
    echo "Requirements:"
    echo "  - git-cliff must be installed"
    echo "  - You must be on main branch with clean working directory"
    echo "  - All tests must pass"
    exit 0
fi

main "$@"
