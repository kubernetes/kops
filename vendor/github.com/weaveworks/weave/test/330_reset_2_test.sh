#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Rolling restart with reset"

weave_on $HOST1 launch-router
weave_on $HOST2 launch-router $HOST1

start_container $HOST1 --name=c1
start_container $HOST2 --name=c2

# Now reset the routers so they should give their space away, and restart
weave_on $HOST1 reset
weave_on $HOST1 launch-router $HOST2
weave_on $HOST2 reset
weave_on $HOST2 launch-router $HOST1

# And try to start a container
assert_raises "timeout 10 cat <( start_container $HOST1 )"

end_suite
