#!/usr/bin/env bash

# Release script for k8s-controller
# This script uses git-cliff to automatically determine version bumps

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

# Check git status
check_git_status() {
    local current_branch=$(git branch --show-current)

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

    local local_commit=$(git rev-parse HEAD)
    local remote_commit=$(git rev-parse origin/main)

    if [ "$local_commit" != "$remote_commit" ]; then
        log_error "Your local main branch is not up to date with origin/main"
        log_info "Please run: git pull origin main"
        exit 1
    fi
}

# Get version info using git-cliff
get_version_info() {
    local current_version=$(git describe --tags --abbrev=0 2>/dev/null || echo "")
    local next_version

    # If no tags exist, git-cliff will use 0.1.0 as default when using --bump
    if [ -z "$current_version" ]; then
        log_info "No existing tags found. This will be the first release."
        current_version="(none)"
        # Use git-cliff to determine initial version
        next_version=$(git cliff --bump --unreleased --context | jq -r '.version // "0.1.0"' 2>/dev/null || echo "0.1.0")
        if [[ ! $next_version =~ ^[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
            next_version="0.1.0"
        fi
        next_version="v${next_version}"
    else
        log_info "Current version: $current_version"
        # Let git-cliff determine the next version based on conventional commits
        next_version=$(git cliff --bump --unreleased --context | jq -r '.version // ""' 2>/dev/null || echo "")
        if [ -z "$next_version" ] || [ "$next_version" = "null" ]; then
            log_warning "No unreleased changes found that warrant a version bump."
            echo "This usually means:"
            echo "  - No conventional commits since last release"
            echo "  - Or only commits that don't trigger version bumps (docs, style, etc.)"
            echo ""
            read -rp "Do you want to force a patch release? (y/N): " force_patch
            if [[ $force_patch == [yY] ]]; then
                # Extract version numbers and increment patch
                local version_nums=${current_version#v}
                local major=${version_nums%%.*}
                local rest=${version_nums#*.}
                local minor=${rest%%.*}
                local patch=${rest##*.}
                next_version="v${major}.${minor}.$((patch + 1))"
            else
                log_info "Release cancelled."
                exit 0
            fi
        else
            next_version="v${next_version}"
        fi
    fi

    echo "Current version: $current_version"
    echo "Next version: $next_version"
    echo ""
}

# Generate changelog and create release
create_release() {
    local next_version="$1"

    log_info "Creating release $next_version..."

    # Run tests to make sure everything is working
    log_info "Running tests..."
    if ! go test ./...; then
        log_error "Tests failed. Please fix them before creating a release."
        exit 1
    fi

    # Generate changelog for the new version
    log_info "Generating changelog..."
    if [ -f "CHANGELOG.md" ]; then
        # Update existing changelog
        git cliff --tag "$next_version" --prepend CHANGELOG.md
    else
        # Create new changelog
        git cliff --tag "$next_version" --output CHANGELOG.md
    fi

    log_success "Changelog generated/updated successfully"

    # Show what will be released
    echo ""
    log_info "Changes to be released in $next_version:"
    echo "----------------------------------------"
    git cliff --tag "$next_version" --unreleased --strip all
    echo "----------------------------------------"
    echo ""

    # Ask for confirmation
    read -rp "Do you want to proceed with creating release $next_version? (y/N): " confirm
    if [[ $confirm != [yY] ]]; then
        log_warning "Release cancelled"
        # Clean up the changelog changes
        git checkout -- CHANGELOG.md 2>/dev/null || true
        exit 0
    fi

    # Commit changelog
    git add CHANGELOG.md
    git commit -m "chore(release): prepare for $next_version"

    # Create and push tag
    git tag -a "$next_version" -m "Release $next_version"

    # Push changes
    log_info "Pushing changes and tag..."
    git push origin main
    git push origin "$next_version"

    log_success "Release $next_version created successfully!"
    log_info "GitHub Actions will now build and publish the release."
    log_info "Check the progress at: https://github.com/$(git config remote.origin.url | sed 's|.*[:/]||' | sed 's|\.git||')/actions"
}

# Main script
main() {
    log_info "Starting release process for k8s-controller..."

    check_git_cliff
    check_git_status

    echo ""
    local version_info
    version_info=$(get_version_info)

    # Extract next version from the output
    local next_version=$(echo "$version_info" | grep "Next version:" | cut -d' ' -f3)

    if [ -z "$next_version" ]; then
        log_error "Could not determine next version"
        exit 1
    fi

    create_release "$next_version"
}

# Show help
if [[ "${1:-}" == "--help" || "${1:-}" == "-h" ]]; then
    echo "Release script for k8s-controller"
    echo ""
    echo "Usage: $0"
    echo ""
    echo "This script will:"
    echo "  1. Check that you're on main branch and up to date"
    echo "  2. Use git-cliff to automatically determine next version"
    echo "  3. Run tests to ensure code quality"
    echo "  4. Generate/update changelog using git-cliff"
    echo "  5. Create a new git tag"
    echo "  6. Push changes and tag to trigger GitHub Actions release"
    echo ""
    echo "Requirements:"
    echo "  - git-cliff must be installed"
    echo "  - You must be on main branch with clean working directory"
    echo "  - All tests must pass"
    echo "  - jq must be installed (for JSON parsing)"
    exit 0
fi

main "$@"
