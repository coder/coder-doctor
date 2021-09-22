#!/usr/bin/env bash
#
# Check dependencies and configuration for security issues

set -euo pipefail
PROJECT_ROOT=$(git rev-parse --show-toplevel)
cd "$PROJECT_ROOT"
source "./scripts/lib.sh"

run_trace false trivy --version

trivy_flags=(
  --vuln-type=os,library
  --severity=MEDIUM,HIGH,CRITICAL
  --exit-code=1
  --security-checks=vuln,config
)

if [ -n "${CI:-}" ]; then
  trivy_flags+=(
    --no-progress
  )
fi

run_trace false trivy filesystem "${trivy_flags[@]}" "$PROJECT_ROOT"
