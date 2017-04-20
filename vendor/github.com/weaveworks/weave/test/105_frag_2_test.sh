#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.1.41
C2=10.2.1.71

# 700 = Linux's minimum PMTU (552) + weave framing overheads
# (currently about 60, but we leave some margin). That way the PMTUs
# seen by containers is still >= 552.
drop_large_sleeve_packets="INPUT -p udp --destination-port 6783 -m length '!' --length 0:700 -j DROP"
drop_large_fastdp_packets="INPUT -p udp --destination-port 6784 -m length '!' --length 0:700 -j DROP"

start_suite "PMTU discovery and packet fragmentation"

weave_on $HOST1 launch
weave_on $HOST2 launch

start_container $HOST1 $C1/24 --name=c1
start_container $HOST2 $C2/24 --name=c2

run_on $HOST1 "sudo iptables -I $drop_large_sleeve_packets"
run_on $HOST1 "sudo iptables -I $drop_large_fastdp_packets"
assert_raises "weave_on $HOST2 connect $HOST1"
sleep 35 # give weave time to discover the PMTU

# Check large packets get through. The first attempt typically fails,
# since the sending container hasn't discovered the PMTU yet. The 2nd
# attempt should succeed.
exec_on $HOST2 c2 $PING -s 10000 $C1 1>/dev/null 2>&1 || true
assert_raises "exec_on $HOST2 c2 $PING -s 10000 $C1"

run_on $HOST1 "sudo iptables -D $drop_large_fastdp_packets"
run_on $HOST1 "sudo iptables -D $drop_large_sleeve_packets"

end_suite
