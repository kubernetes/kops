#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "IPAM alloc range change"

weave_on $HOST1 launch --ipalloc-range 10.2.0.0/16
weave_on $HOST1 prime
weave_on $HOST1 stop
weave_on $HOST1 launch --ipalloc-range 10.3.0.0/16
# Ensure allocations can proceed
assert_raises "timeout 10 cat <( start_container $HOST1 )"

end_suite
