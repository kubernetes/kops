#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

# Kill child processes at exit
trap "pkill -P $$" SIGINT SIGTERM EXIT

go install github.com/weaveworks/mesh/metcd/metcdsrv

metcdsrv -quicktest=1 &
metcdsrv -quicktest=2 &
metcdsrv -quicktest=3 &

read x

