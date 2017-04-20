#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.1.4
C2=10.2.1.7
UNIVERSE=10.2.0.0/16
SUBNET_1=10.2.2.0/24
SUBNET_2=10.2.3.0/24

start_suite "Ping over cross-host weave network (with and without IPAM)"

weave_on $HOST1 launch --ipalloc-range $UNIVERSE --ipalloc-default-subnet $SUBNET_1
weave_on $HOST2 launch --ipalloc-range $UNIVERSE --ipalloc-default-subnet $SUBNET_1 $HOST1

start_container $HOST1    $C1/24     --name=c1
start_container $HOST2 ip:$C2/24     --name=c2
assert_raises "exec_on $HOST1 c1 $PING $C2"

start_container $HOST1               --name=c3
start_container $HOST2 net:default   --name=c4
C4=$(container_ip $HOST2 c4)
assert_raises "exec_on $HOST1 c3 $PING $C4"

start_container $HOST1 net:$SUBNET_2 --name=c5
start_container $HOST2 net:$SUBNET_2 --name=c6
C6=$(container_ip $HOST2 c6)
assert_raises "exec_on $HOST1 c5 $PING $C6"

# check that restart+attach retains the same IP, and reclaims it in IPAM
docker_on $HOST2 restart c6
weave_on $HOST2 attach net:$SUBNET_2 c6
assert_raises "exec_on $HOST1 c5 $PING $C6"
# check that restart does not create any additional DNS entry
assert "weave_on $HOST2 report -f \"{{len .DNS.Entries}}\"" "6"
sleep 31 # past the IPAM timeout
start_container $HOST2 net:$SUBNET_2 --name=c7
C7=$(container_ip $HOST2 c7)
assert_raises "[ $C6 != $C7 ]"

end_suite
