#!/usr/bin/env bash
#
# This script installs dependencies to /usr/local/bin.

set -euo pipefail
PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "$PROJECT_ROOT"
source "./scripts/lib.sh"

TMPDIR=$(mktemp -d)
TMPBIN="${TMPDIR}/bin"
BINDIR="/usr/local/bin"

curl_flags=(
  --silent
  --show-error
  --location
)

# Install Go programs
export GOPATH="$TMPDIR/go"

run_trace false mkdir --parents "$GOPATH"

# goveralls collects code coverage metrics from tests
# and sends to Coveralls
run_trace false go install github.com/mattn/goveralls@v0.0.9

# Install binaries only
run_trace false sudo install --mode=0755 --target-directory="$BINDIR" "$GOPATH/bin/*"

# Install packages via apt where available
run_trace false sudo apt-get install --no-install-recommends --yes \
  shellcheck

# Extract binaries as non-root user, then sudo install
run_trace false mkdir --parents "$TMPBIN"

# gotestsum makes test output more readable
GOTESTSUM_VERSION="1.7.0"
run_trace false curl "${curl_flags[@]}" "https://github.com/gotestyourself/gotestsum/releases/download/v${GOTESTSUM_VERSION}/gotestsum_${GOTESTSUM_VERSION}_linux_amd64.tar.gz" \| \
  tar --extract --gzip --directory="$TMPBIN" --file=- gotestsum

# golangci-lint to lint Go code with multiple tools
GOLANGCI_LINT_VERSION="1.42.1"
run_trace false curl "${curl_flags[@]}" "https://github.com/golangci/golangci-lint/releases/download/v${GOLANGCI_LINT_VERSION}/golangci-lint-${GOLANGCI_LINT_VERSION}-linux-amd64.tar.gz" \| \
  tar --extract --gzip --directory="$TMPBIN" --file=- --strip-components=1 "golangci-lint-${GOLANGCI_LINT_VERSION}-linux-amd64/golangci-lint"

# goreleaser to compile, cross-compile, and release binaries
GORELEASER_VERSION="0.178.0"
run_trace false curl "${curl_flags[@]}" "https://github.com/goreleaser/goreleaser/releases/download/v${GORELEASER_VERSION}/goreleaser_Linux_x86_64.tar.gz" \| \
  tar --extract --gzip --directory="$TMPBIN" --file=- "goreleaser"

# trivy to scan container images
TRIVY_VERSION="0.19.2"
run_trace false curl "${curl_flags[@]}" "https://github.com/aquasecurity/trivy/releases/download/v${TRIVY_VERSION}/trivy_${TRIVY_VERSION}_Linux-64bit.tar.gz" \| \
  tar --extract --gzip --directory="$TMPBIN" --file=- "trivy"

run_trace false sudo install --mode=0755 --target-directory="$BINDIR" "$TMPBIN/*"

run_trace false command -v \
  golangci-lint \
  goreleaser \
  gotestsum \
  trivy

run_trace false sudo rm --verbose --recursive --force "$TMPDIR"
