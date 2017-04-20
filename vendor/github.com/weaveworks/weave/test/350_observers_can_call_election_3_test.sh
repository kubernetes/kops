#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Observers can call election"

# Start network with sufficient voters for consensus
weave_on $HOST1 launch --ipalloc-init consensus=2
weave_on $HOST2 launch --ipalloc-init consensus=2 $HOST1

# Add an observer
weave_on $HOST3 launch --ipalloc-init observer $HOST2

# Wait for network to settle; ensure IPAM is idle (blank output = no consensus)
sleep 5
assert "weave_on $HOST1 status ipam" ""
assert "weave_on $HOST2 status ipam" ""
assert "weave_on $HOST3 status ipam" ""

# Check allocation succeeds on observer
assert_raises "timeout 10 cat <( start_container $HOST3 )"

end_suite
