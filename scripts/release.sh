#!/bin/bash

# GitOps Validator Release Script
# Usage: ./scripts/release.sh [version] [message]
# Example: ./scripts/release.sh v1.0.0 "Initial release"

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if we're in a git repository
check_git_repo() {
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        print_error "Not in a git repository!"
        exit 1
    fi
}

# Function to check if there are uncommitted changes
check_uncommitted_changes() {
    if ! git diff-index --quiet HEAD --; then
        print_warning "You have uncommitted changes!"
        echo "Do you want to continue anyway? (y/N)"
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_error "Release cancelled."
            exit 1
        fi
    fi
}

# Function to check if tag already exists
check_tag_exists() {
    local version=$1
    if git rev-parse "$version" >/dev/null 2>&1; then
        print_error "Tag $version already exists!"
        exit 1
    fi
}

# Function to validate version format
validate_version() {
    local version=$1
    if [[ ! $version =~ ^v[0-9]+\.[0-9]+\.[0-9]+$ ]]; then
        print_error "Invalid version format! Use format: v1.0.0"
        exit 1
    fi
}

# Function to update version in files (if needed)
update_version() {
    local version=$1
    print_status "Updating version to $version..."
    
    # Update version in main.go if it exists
    if [ -f "main.go" ]; then
        # This is a placeholder - you might want to update version constants
        print_status "Version $version ready for release"
    fi
}

# Function to create release
create_release() {
    local version=$1
    local message=$2
    
    print_status "Creating release $version..."
    
    # Add all changes
    git add .
    
    # Commit changes
    git commit -m "Prepare for release $version" || print_warning "No changes to commit"
    
    # Create tag
    git tag -a "$version" -m "$message"
    
    # Push changes and tag
    git push origin main
    git push origin "$version"
    
    print_success "Release $version created and pushed!"
    print_status "GitHub Actions will now build and upload the release assets."
    print_status "Check the Actions tab in your GitHub repository to monitor progress."
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [version] [message]"
    echo ""
    echo "Examples:"
    echo "  $0 v1.0.0 \"Initial release\""
    echo "  $0 v1.1.0 \"Bug fixes and improvements\""
    echo "  $0 v2.0.0 \"Major release with new features\""
    echo ""
    echo "Version format: vX.Y.Z (e.g., v1.0.0)"
    echo "Message: Description of the release"
}

# Main function
main() {
    # Check if help is requested
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        show_usage
        exit 0
    fi
    
    # Check arguments
    if [ $# -lt 2 ]; then
        print_error "Missing required arguments!"
        echo ""
        show_usage
        exit 1
    fi
    
    local version=$1
    local message=$2
    
    print_status "Starting release process for $version..."
    
    # Validate inputs
    validate_version "$version"
    
    # Check prerequisites
    check_git_repo
    check_uncommitted_changes
    check_tag_exists "$version"
    
    # Update version
    update_version "$version"
    
    # Create release
    create_release "$version" "$message"
    
    print_success "Release process completed!"
    print_status "Release URL: https://github.com/$(git config --get remote.origin.url | sed 's/.*github.com[:/]\([^.]*\).*/\1/')/releases/tag/$version"
}

# Run main function with all arguments
main "$@"
