#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "IPAM ring seed mismatch"

weave_on $HOST1 launch-router --no-discovery
weave_on $HOST2 launch-router --no-discovery
weave_on $HOST3 launch-router --no-discovery

# Do some allocations to get the ring initialized on two hosts
start_container $HOST1
start_container $HOST2

# Connect 2 to 3 and allocate so they will each have half the ring
weave_on $HOST2 connect $HOST3
start_container $HOST3

# More allocations on 1 to increase the version number
start_container $HOST1
start_container $HOST1

weave_on $HOST1 connect $HOST3

# Check that weave on 3 is still running
assert_raises "weave_on $HOST3 status"

end_suite
