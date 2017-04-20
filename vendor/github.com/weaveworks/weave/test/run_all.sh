#!/bin/bash

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
. "$DIR/config.sh"

whitely echo Sanity checks
if ! bash "$DIR/sanity_check.sh"; then
    whitely echo ...failed
    exit 1
fi
whitely echo ...ok

TESTS="${@:-$(find "$DIR" -name '*_test.sh')}"
RUNNER_ARGS=${RUNNER_ARGS:-""}

# If running on circle, use the scheduler to work out what tests to run
if [ -n "$CIRCLECI" -a -z "$NO_SCHEDULER" ]; then
    RUNNER_ARGS="$RUNNER_ARGS -scheduler"
fi

# If running on circle or PARALLEL is not empty, run tests in parallel
if [ -n "$CIRCLECI" -o -n "$PARALLEL" ]; then
    RUNNER_ARGS="$RUNNER_ARGS -parallel"
fi

HOSTS="$HOSTS" "${DIR}/../tools/runner/runner" $RUNNER_ARGS $TESTS
