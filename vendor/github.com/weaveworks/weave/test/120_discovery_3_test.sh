#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.1.41
C3=10.2.1.71

launch_all() {
    weave_on $HOST1 launch-router $1
    weave_on $HOST2 launch-router $1 --ipalloc-range="" $HOST1
    weave_on $HOST3 launch-router $1                    $HOST2
}

start_suite "Peer discovery, multi-hop routing and gossip forwarding"

launch_all

start_container $HOST1 $C1/24 --name=c1
start_container $HOST3 $C3/24 --name=c3

assert_raises "exec_on $HOST1 c1 $PING $C3"
stop_weave_on $HOST2
assert_raises "exec_on $HOST1 c1 $PING $C3"

stop_weave_on $HOST1
stop_weave_on $HOST3

launch_all --no-discovery

sleep 5 # give topology gossip some time to propagate

assert_raises "exec_on $HOST1 c1 $PING $C3"

assert_raises "start_container $HOST1" # triggers IPAM initialisation
# this stalls if gossip forwarding doesn't work. We wait for slightly
# longer than the gossip interval (30s) before giving up.
assert_raises "timeout 40 cat <( start_container $HOST3 )"

stop_weave_on $HOST2
assert_raises "exec_on $HOST1 c1 sh -c '! $PING $C3'"

end_suite
