#! /bin/bash

. "$(dirname "$0")/config.sh"

# rot8 the hex digits so if you do this twice you get back the original
rot8() {
    tr '[0-9a-f]' '[8-9a-f0-7]'
}

get_peername() {
    weave_on $HOST1 report -f '{{.Router.Name}}'
}

remove_bridge() {
    run_on $HOST1 sudo ip link del weave
    run_on $HOST1 sudo ip link del vethwe-datapath
    docker_on $HOST1 run --rm --privileged --net=host --entrypoint=/usr/bin/weaveutil weaveworks/weaveexec delete-datapath datapath
}

start_suite "peer name derivation"

# Save the original machine ID
ORIG_MACHINE_ID=$($SSH $HOST1 cat /etc/machine-id)
[ -n "$ORIG_MACHINE_ID" ] || echo "Machine ID is blank!" >&2

# Blank machine-id to begin, so peername will work the same as pre-1.9
$SSH $HOST1 sudo cp /dev/null /etc/machine-id

weave_on $HOST1 launch-router
weave_on $HOST1 prime
PEERNAME1=$(get_peername)

# Now put the machine ID back
echo $ORIG_MACHINE_ID | $SSH $HOST1 sudo tee /etc/machine-id >/dev/null

# Stop and restart; it should take the ID from the bridge
weave_on $HOST1 stop-router
weave_on $HOST1 launch-router

assert "get_peername" "$PEERNAME1"

# Stop and remove the bridge, then restart; it should take the ID from persistence
weave_on $HOST1 stop-router
remove_bridge
weave_on $HOST1 launch-router

assert "get_peername" "$PEERNAME1"

# Now remove everything and restart: should get a different peer name
weave_on $HOST1 reset
weave_on $HOST1 launch-router

PEERNAME3=$(get_peername)
assert_raises "[ $PEERNAME3 != $PEERNAME1 ]"

# Change the machine ID
echo $ORIG_MACHINE_ID | rot8 | $SSH $HOST1 sudo tee /etc/machine-id >/dev/null

# Remove bridge and restart; it should get a different ID
weave_on $HOST1 stop-router
remove_bridge
weave_on $HOST1 launch-router

assert_raises "[ $(get_peername) != $PEERNAME3 ]"

# Put the original machine ID back
echo $ORIG_MACHINE_ID | $SSH $HOST1 sudo tee /etc/machine-id >/dev/null

end_suite
