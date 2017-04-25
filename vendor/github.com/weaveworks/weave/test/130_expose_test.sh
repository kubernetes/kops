#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.1.34
C2=10.2.1.37
C3=10.2.2.34
C4=10.2.2.37
EXP=10.2.2.101
UNIVERSE=10.2.3.0/24

weave_on1() {
    assert_raises "weave_on $HOST1 $@"
}

run_on1() {
    assert_raises "run_on   $HOST1 $@"
}

exec_on1() {
    assert_raises "exec_on  $HOST1 $@"
}

# Containers in the same subnet should be able to talk; different subnet not.
check_container_connectivity() {
    exec_on1 "c1 $PING $C2"
    exec_on1 "c3 $PING $C4"
    exec_on1 "c5 $PING $C6"
    exec_on1 "c1 sh -c '! $PING $C3'"
    exec_on1 "c3 sh -c '! $PING $C5'"
    exec_on1 "c5 sh -c '! $PING $C1'"
}

start_suite "exposing weave network to host"

# expose/hide with a CIDR work prior to launching weave
run_on1   "! $PING $EXP"
weave_on1 "expose  $EXP/24"
run_on1   "  $PING $EXP"
weave_on1 "hide    $EXP/24"
run_on1   "! $PING $EXP"

weave_on $HOST1 launch --ipalloc-range $UNIVERSE

start_container $HOST1 $C1/24 --name=c1
start_container $HOST1 $C2/24 --name=c2
start_container $HOST1 $C3/24 --name=c3
start_container $HOST1 $C4/24 --name=c4
start_container $HOST1        --name=c5
start_container $HOST1        --name=c6
C5=$(container_ip $HOST1 c5)
C6=$(container_ip $HOST1 c6)

# absence of host connectivity by default
run_on1 "! $PING $C3"
run_on1 "! $PING $C5"
check_container_connectivity

# host connectivity after 'expose', and idempotence of 'expose'
weave_on1 "expose $EXP/24"
weave_on1 "expose $EXP/24"
run_on1   "! $PING $C1"
run_on1   "  $PING $C3"
run_on1   "! $PING $C5"
weave_on1 "expose"
weave_on1 "expose"
run_on1   "! $PING $C1"
run_on1   "  $PING $C3"
run_on1   "  $PING $C5"
check_container_connectivity

# no host connectivity after 'hide', and idempotence of 'hide'
weave_on1 "hide $EXP/24"
weave_on1 "hide $EXP/24"
run_on1   "! $PING $C3"
run_on1   "  $PING $C5"
weave_on1 "hide"
weave_on1 "hide"
run_on1   "! $PING $C5"

end_suite
