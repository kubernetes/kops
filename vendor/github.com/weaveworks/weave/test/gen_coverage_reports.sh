#!/bin/bash

set -ex
DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

if [ -n "$CIRCLECI" ]; then
    for i in $(seq 1 $(($CIRCLE_NODE_TOTAL - 1))); do
        scp node$i:/home/ubuntu/src/github.com/weaveworks/weave/test/coverage/* ./coverage/ || true
    done
fi

go get github.com/weaveworks/build-tools/cover
cover ./coverage/* >profile.cov
go tool cover -html=profile.cov -o coverage.html
go tool cover -func=profile.cov -o coverage.txt
tar czf coverage.tar.gz ./coverage
