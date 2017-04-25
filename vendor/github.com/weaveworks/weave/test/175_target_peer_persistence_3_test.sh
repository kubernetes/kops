#! /bin/bash

. "$(dirname "$0")/config.sh"


assert_targets() {
    HOST=$1
    shift
    EXPECTED=$(for TARGET in $@; do echo $TARGET; done | sort)
    assert "weave_on $HOST report -f '{{range .Router.Targets}}{{.}}{{\"\n\"}}{{end}}' | sort" "$EXPECTED"
}

start_suite "Check Docker restart uses persisted peer list"

# Launch router and modify initial peer list
weave_on $HOST1 launch $HOST1 $HOST2
weave_on $HOST1 forget $HOST1
weave_on $HOST1 connect $HOST3

# Ensure modified peer list is still in effect after restart
check_restart $HOST1 weave
assert_targets $HOST1 $HOST2 $HOST3

# Ensure persisted peer changes are still in effect after --resume
weave_on $HOST1 stop
weave_on $HOST1 launch --resume
assert_targets $HOST1 $HOST2 $HOST3

# Ensure persisted peer changes are ignored after stop and subsequent restart
weave_on $HOST1 stop
weave_on $HOST1 launch $HOST1 $HOST2
assert_targets $HOST1 $HOST1 $HOST2
check_restart $HOST1 weave
assert_targets $HOST1 $HOST1 $HOST2

end_suite
