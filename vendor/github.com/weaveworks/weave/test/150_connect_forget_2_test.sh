#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.1.4
C2=10.2.1.7

direct_peers() {
    weave_on $1 report -f "{{range .Router.Targets}}{{.}}
{{end}}"
}

assert_peers() {
  assert "direct_peers $1 | sort" "$(echo -e "$2" | sort)"
}

start_suite "Connecting and forgetting routers after launch"

weave_on $HOST1 launch-router
weave_on $HOST2 launch-router

start_container $HOST1 $C1/24 --name=c1
start_container $HOST2 $C2/24 --name=c2

# Connect to some bogus hosts
assert_raises "weave_on $HOST2 connect 10.2.2.1 10.2.2.2"
assert_peers $HOST2 "10.2.2.1\n10.2.2.2"

# Replace bogus hosts with real hosts
assert_raises "exec_on $HOST1 c1 sh -c '! $PING $C2'"
assert_raises "weave_on $HOST2 connect --replace $HOST1 $HOST2"
assert_peers $HOST2 "$HOST1\n$HOST2"
assert_raises "exec_on $HOST1 c1 $PING $C2"

# Forget everyone and disconnect
assert_raises "weave_on $HOST2 forget $HOST1 $HOST2"
assert_raises "stop_weave_on $HOST1"
assert_raises "weave_on $HOST1 launch-router"
assert_peers $HOST2 ""
assert_raises "exec_on $HOST1 c1 sh -c '! $PING $C2'"

end_suite
