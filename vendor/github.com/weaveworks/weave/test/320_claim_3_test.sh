#! /bin/bash

. "$(dirname "$0")/config.sh"

UNIVERSE=10.2.0.0/16
C1=10.2.128.1
C4=10.2.128.4

delete_persistence() {
    for host in "$@" ; do
        docker_on $host rm -v weavedb >/dev/null
        docker_on $host rm weave >/dev/null
    done
}

start_suite "Claiming addresses"

weave_on $HOST1 launch-router --ipalloc-range $UNIVERSE $HOST2
weave_on $HOST2 launch-router --ipalloc-range $UNIVERSE $HOST1

start_container $HOST1 $C1/12 --name=c1
start_container $HOST2        --name=c2
assert_raises "exec_on $HOST2 c2 $PING $C1"

stop_weave_on $HOST1
stop_weave_on $HOST2

# Delete persistence data so they form a blank ring
delete_persistence $HOST1 $HOST2

# Start hosts in reverse order so c1's address has to be claimed from host2
weave_on $HOST2 launch-router --ipalloc-range $UNIVERSE
weave_on $HOST1 launch-router --ipalloc-range $UNIVERSE $HOST2

# Start another container on host2, so if it hasn't relinquished c1's
# address it would give that out as the first available.
start_container $HOST2 --name=c3
C3=$(container_ip $HOST2 c3)
assert_raises "[ $C3 != $C1 ]"

sleep 1 # give routers some time to fully establish connectivity
assert_raises "exec_on $HOST1 c1 $PING $C3"

stop_weave_on $HOST1
stop_weave_on $HOST2

delete_persistence $HOST1 $HOST2

# Now make host1 attempt to claim from host2, when host2 is stopped
# the point being to check whether host1 will hang trying to talk to host2
weave_on $HOST2 launch-router --ipalloc-range $UNIVERSE
weave_on $HOST2 prime
# Introduce host3 to remember the IPAM CRDT when we stop host2
weave_on $HOST3 launch-router --ipalloc-range $UNIVERSE $HOST2
weave_on $HOST3 prime
stop_weave_on $HOST2
weave_on $HOST1 launch-router --ipalloc-range $UNIVERSE $HOST3

stop_weave_on $HOST1
stop_weave_on $HOST3

# Delete persistence data on host1, so that host1 would try to establish a ring.
delete_persistence $HOST1
weave_on $HOST1 launch --ipalloc-range $UNIVERSE $HOST2

# Start another container on host1. The starting should block, because host1 is
# not able to establish the ring due to host2 being offline.
CMD="proxy_start_container $HOST1 --name c4 -e WEAVE_CIDR=$C4/12"
assert_raises "timeout 5 cat <( $CMD )" 124

# However, allocation for an external subnet should not block.
assert_raises "proxy_start_container $HOST1 -e WEAVE_CIDR=10.3.0.1/12"

# Launch host2, so that host1 can establish the ring.
weave_on $HOST2 launch --ipalloc-range $UNIVERSE $HOST1
wait_for_attached $HOST1 c4
assert_raises "[ $(container_ip $HOST1 c4) == $C4 ]"

end_suite
