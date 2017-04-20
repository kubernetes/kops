#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Specify initial IPAM seed"

# Launch two disconnected routers
weave_on $HOST1 launch --name ::1 --ipalloc-init seed=::1,::2
weave_on $HOST2 launch --name ::2 --ipalloc-init seed=::1,::2

# Ensure allocations can proceed
assert_raises "timeout 10 cat <( start_container $HOST1 --name c1)"
assert_raises "timeout 10 cat <( start_container $HOST2 --name c2)"

# Connect routers
weave_on $HOST2 connect $HOST1

# Check connectivity
assert_raises "exec_on $HOST1 c1 $PING c2"
assert_raises "exec_on $HOST2 c2 $PING c1"

# Now restart one router with a different peername
docker_on $HOST2 rm -f c2 >/dev/null
weave_on $HOST2 forget $HOST1
weave_on $HOST2 stop
weave_on $HOST2 launch --name ::3
# Check that this host has not retained the previous IPAM data
assert "weave_on $HOST2 status ipam" ""

end_suite
