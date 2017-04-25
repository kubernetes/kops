#! /bin/bash

. "$(dirname "$0")/config.sh"

C1_NAME=c1.weave.local

start_suite "Recovery of container details on router restart (via 'local' plugin)"

weave_on $HOST1 launch
start_container_local_plugin $HOST1 --name=c1 --hostname=$C1_NAME
c1ip=$(container_ip $HOST1 c1)

weave_on $HOST1 stop-router
weave_on $HOST1 launch-router

assert "container_ip $HOST1 c1" "$c1ip"
assert "weave_on $HOST1 dns-lookup $C1_NAME" "$c1ip"

# check that c1 IP has been reclaimed and doesn't get assigned to a
# fresh container
start_container $HOST1 --name=c2
assert_raises "container_ip $HOST1 c2 | grep -v $c1ip"

docker_on $HOST1 rm -f c1 c2

end_suite
