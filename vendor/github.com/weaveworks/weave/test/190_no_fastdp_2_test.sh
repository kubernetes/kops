#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.1.46
C2=10.2.1.47

start_suite "WEAVE_NO_FASTDP operation"

WEAVE_NO_FASTDP=1 weave_on $HOST1 launch
WEAVE_NO_FASTDP=1 weave_on $HOST2 launch $HOST1

start_container $HOST1 $C1/24 --name=c1
start_container $HOST2 $C2/24 --name=c2

# Without fastdp, ethwe should have an MTU of 65535
assert_raises "exec_on $HOST1 c1 sh -c 'ip link show ethwe | grep \ mtu\ 65535\ >/dev/null'"

# Pump some data over a TCP socket between the containers.  This
# should cause PMTU discovery on the overlay network, and hence
# exercise the sleeve code to generate an ICMP frag-needed.
assert_raises "exec_on $HOST1 c1 sh -c 'nc -l -p 5555 </dev/null >/tmp/foo'" &
sleep 5
assert_raises "exec_on $HOST2 c2 sh -c 'dd if=/dev/zero bs=10k count=1 | nc $C1 5555'"
assert_raises "exec_on $HOST1 c1 sh -c 'test \$(stat -c %s /tmp/foo) -eq 10240'"

end_suite
