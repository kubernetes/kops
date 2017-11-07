#!/usr/bin/env bash

set -x -e

if [ "$TEST" == "example" ]; then
	cd example
	PATH=$PATH:/tmp/test-etcd make test
elif [ "$TEST" == "test" ]; then
	cd test
	PATH=$PATH:/tmp/test-etcd:`pwd`/bin/ make test
fi
