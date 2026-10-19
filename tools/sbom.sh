#!/usr/bin/env bash
# Generates SBOM from the bazel binary.
# The version comes from BUILD.bazel in role pkg_deb.
set -euo pipefail

TOOL="prometheus-mgit-exporter"
TARGET="//:prometheus-mgit-exporter"
BUILD_FILE="BUILD.bazel"
CDX_VERSION="${CDX_VERSION:-v1.12.0}"

VERSION="$(sed -n 's/.*version *= *"\([^"]*\)".*/\1/p' "$BUILD_FILE" | head -n1)"
[[ -n "$VERSION" ]] || { echo "Keine version in $BUILD_FILE gefunden" >&2; exit 1; }

GOROOT="$(bazel run @rules_go//go -- env GOROOT 2>/dev/null)"
export PATH="${GOROOT}/bin:${PATH}"
go install "github.com/CycloneDX/cyclonedx-gomod/cmd/cyclonedx-gomod@${CDX_VERSION}"
export PATH="$PATH:$(go env GOPATH)/bin"

bazel build "$TARGET"
bin="$(bazel info execution_root 2>/dev/null)/$(bazel cquery --output=files "$TARGET" 2>/dev/null)"

mkdir -p sbom
cyclonedx-gomod bin -json -licenses -assert-licenses -std -version "$VERSION" -output "sbom/${TOOL}.cdx.json" "$bin"
echo "SBOM_RELEASE_VERSION=${VERSION}" > sbom/upload.env
