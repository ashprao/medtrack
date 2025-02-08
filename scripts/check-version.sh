#!/bin/bash

# Exit on any error
set -e

# Get the version from internal/version/version.go
VERSION=$(grep 'Version = ' internal/version/version.go | cut -d'"' -f2)
echo "Version from version.go: $VERSION"

# Get the latest git tag
LATEST_TAG=$(git describe --tags --abbrev=0 2>/dev/null || echo "no tags found")
echo "Latest git tag: $LATEST_TAG"

# Check if version.go matches git tag
if [ "$LATEST_TAG" != "v$VERSION" ] && [ "$LATEST_TAG" != "no tags found" ]; then
    echo "Error: Version mismatch!"
    echo "version.go has $VERSION but git tag is $LATEST_TAG"
    exit 1
fi

# Check CHANGELOG.md
if ! grep -q "## \[v$VERSION\]" CHANGELOG.md; then
    echo "Error: Version $VERSION not found in CHANGELOG.md"
    exit 1
fi

# Check README.md installation instructions
if ! grep -q "go install github.com/ashprao/medtrack/cmd/medtrack@v$VERSION" README.md; then
    echo "Error: Installation instructions in README.md don't match version $VERSION"
    exit 1
fi

echo "✅ Version $VERSION is consistent across all files"
