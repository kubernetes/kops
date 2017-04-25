#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Observers can't vote"

# Start network that doesn't have enough peers for consensus
weave_on $HOST1 launch --ipalloc-init consensus=2
weave_on $HOST2 launch --ipalloc-init observer $HOST1

# Check that an allocation times out
assert_raises "timeout 2 cat <( start_container $HOST1 )" 124

# Add another elector
weave_on $HOST3	launch --ipalloc-init consensus=2 $HOST2

# Check allocation succeeds
assert_raises "timeout 10 cat <( start_container $HOST1 )"

end_suite
