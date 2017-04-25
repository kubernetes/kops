#!/usr/bin/env bash

# This is just a sanity check for metcdsrv.

set -o errexit
set -o nounset
set -o pipefail

# Kill child processes at exit
trap "pkill -P $$" SIGINT SIGTERM EXIT

echo Installing metcdsrv
go install github.com/weaveworks/mesh/metcd/metcdsrv

echo Booting cluster
# Remove output redirection to debug
metcdsrv -quicktest=1 >/dev/null 2>&1 &
metcdsrv -quicktest=2 >/dev/null 2>&1 &
metcdsrv -quicktest=3 >/dev/null 2>&1 &

echo Waiting for cluster to settle
# Wait for the cluster to settle
sleep 5

echo Installing etcdctl
go install github.com/coreos/etcd/cmd/etcdctl
function etcdctl { env ETCDCTL_API=3 etcdctl --endpoints=127.0.0.1:8001,127.0.0.1:8002,127.0.0.1:8003 $*; }

echo Testing first put
etcdctl put foo bar
have=$(etcdctl get foo | tail -n1)
want="bar"
if [[ $want != $have ]]
then
	echo foo: want $want, have $have
	exit 1
fi

echo Testing second put
etcdctl put foo baz
have=$(etcdctl get foo | tail -n1)
want="baz"
if [[ $want != $have ]]
then
	echo foo: want $want, have $have
	exit 1
fi
