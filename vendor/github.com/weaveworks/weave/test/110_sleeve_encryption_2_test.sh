#!/bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.1.4
C2=10.2.1.7

start_suite "Ping over encrypted cross-host weave network (sleeve)"

WEAVE_NO_FASTDP=1 weave_on $HOST1 launch --password wfvAwt7sj
WEAVE_NO_FASTDP=1 weave_on $HOST2 launch --password wfvAwt7sj $HOST1

start_container $HOST1 $C1/24 --name=c1
start_container $HOST2 $C2/24 --name=c2
assert_raises "exec_on $HOST1 c1 $PING $C2"

assert_raises "weave_on $HOST1 status connections | grep -P 'encrypted *sleeve'"
assert_raises "weave_on $HOST2 status connections | grep -P 'encrypted *sleeve'"

end_suite
