#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.1.43
C2=10.2.1.44
C3=10.2.1.45

start_suite "re-using IP addresses (arp cache updating)"

weave_on $HOST1 launch

check() {
    assert_raises "exec_on $HOST1 c1 $PING $C2"
}

start_container $HOST1 $C1/24 --name=c1

start_container $HOST1 $C2/24 --name=c2
check

docker_on $HOST1 stop -t 1 c2
start_container $HOST1 $C2/24 --name=c3
check

docker_on $HOST1 stop -t 1 c3
start_container $HOST1 $C3/24 --name=c4
weave_on $HOST1 attach $C2/24 c4
check

end_suite
