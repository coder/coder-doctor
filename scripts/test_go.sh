#!/usr/bin/env bash
#
# Run unit and integration tests for Go code

set -euo pipefail
PROJECT_ROOT="$(git rev-parse --show-toplevel)"
cd "$PROJECT_ROOT"
source "./scripts/lib.sh"

echo "--- Running go test"
export FORCE_COLOR=true

test_args=(
  -v
  -failfast
  "${TEST_ARGS:-}"
)

REPORTDIR="/tmp/testreports"
mkdir -p "$REPORTDIR"
TESTREPORT_JSON="$REPORTDIR/test_go.json"
TESTREPORT_XML="$REPORTDIR/test_go.xml"
COVERAGE="$REPORTDIR/test_go.coverage"

test_args+=(
  "-covermode=set"
  "-coverprofile=$COVERAGE"
)

# Allow failures to ensure that we can upload coverage
set +e

run_trace false gotestsum \
  --debug \
  --jsonfile="$TESTREPORT_JSON" \
  --junitfile="$TESTREPORT_XML" \
  --hide-summary=skipped \
  --packages="./..." \
  -- "${test_args[@]}"
test_status=$?

# The following steps should never fail, and if they do,
# we want to know about it, so fail the build
set -e

threshold="5s"
echo "--- ಠ_ಠ The following tests took longer than $threshold to complete:"
run_trace false gotestsum tool slowest \
  --jsonfile="$TESTREPORT_JSON" \
  --threshold="$threshold"

# From time to time, Coveralls seems to have an issue on their end, so
# make a best-effort attempt to upload coverage but ignore failures
set +e
echo "--- Uploading test coverage report to Coveralls..."
run_trace false goveralls -service=github -coverprofile="$COVERAGE"
set -e

exit $test_status
