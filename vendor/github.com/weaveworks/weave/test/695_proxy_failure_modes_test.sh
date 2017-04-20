#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Proxy failure modes"

# docker run should fail when weave router is not running
weave_on $HOST1 launch-proxy
assert_raises "! proxy docker_on $HOST1 run --rm $SMALL_IMAGE true"
assert_raises "proxy docker_on $HOST1 run --rm $SMALL_IMAGE true 2>&1 1>/dev/null | grep 'Error response from daemon: weave container is not present. Have you launched it?'"

end_suite
