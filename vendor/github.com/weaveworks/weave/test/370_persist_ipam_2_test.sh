#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Checking persistence of IPAM"

launch_router_with_db() {
    host=$1
    shift
    weave_on $host launch-router "$@"
}

launch_router_with_db $HOST1 $HOST2
launch_router_with_db $HOST2 $HOST1

EXPOSE=$(weave_on $HOST1 expose)
start_container $HOST1 --name=c1
C1=$(container_ip $HOST1 c1)
start_container $HOST2 --name=c2
assert_raises "exec_on $HOST2 c2 $PING $C1"
start_container $HOST2 --name=c3a
C3a=$(container_ip $HOST2 c3a)

stop_weave_on $HOST1
stop_weave_on $HOST2

# Stop one container while weave is down; the address should be freed on launch
docker_on $HOST2 rm -f c3a

# Start just HOST2; if nothing persisted it would form its own ring
launch_router_with_db $HOST2
start_container $HOST2 --name=c3
C3=$(container_ip $HOST2 c3)
assert_raises "[ $C3 != $C1 ]"
assert_raises "[ $C3 = $C3a ]"

# Restart HOST1 and see if it remembers to connect to HOST2
#
# disabled pending rationalisation of the expected behaviour
#
# launch_router_with_db $HOST1
# assert_raises "exec_on $HOST2 c2 $PING $C1"

# Restart the router without going through the reclaim procedure on launch
docker_on $HOST1 restart weave
# Check that weave:expose address hasn't been lost
start_container $HOST1 --name=c4
C4=$(container_ip $HOST1 c4)
assert_raises "[ $C4 != $EXPOSE ]"

end_suite
