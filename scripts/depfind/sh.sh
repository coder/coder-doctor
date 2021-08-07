#!/usr/bin/env bash

builtin set -euo pipefail
PROJECT_ROOT="$(git rev-parse --show-toplevel)"

builtin pushd "$PROJECT_ROOT" > /dev/null
  git ls-files --full-name -- \
      '.husky' \
      '*.sh' | \
    xargs -IX echo "$PROJECT_ROOT/X"
builtin popd > /dev/null
