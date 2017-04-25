#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Detect hairpin mode"

weave_on $HOST1 launch

assert_raises "run_on $HOST1 sudo ip link set vethwe-bridge type bridge_slave hairpin on"
assert_raises "docker_on $HOST1 logs weave 2>&1 | grep -q 'Hairpin mode enabled on \"vethwe-bridge\"'"

end_suite
