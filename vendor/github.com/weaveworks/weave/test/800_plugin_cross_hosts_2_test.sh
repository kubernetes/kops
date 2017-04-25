#! /bin/bash

. "$(dirname "$0")/config.sh"

C1_NAME=c1.weave.local
C2_NAME=seetwo.weave.local

start_suite "Ping and DNS over cross-host weave network (via 'local' plugin)"

DNS_IP=$(weave_on $host docker-bridge-ip)

weave_on $HOST1 launch
weave_on $HOST2 launch $HOST1

start_container_local_plugin $HOST1 --name=c1 --hostname=$C1_NAME --dns=$DNS_IP
start_container_local_plugin $HOST2 --name=c2 --hostname=$C2_NAME --dns=$DNS_IP

assert_raises "exec_on $HOST1 c1 $PING $C2_NAME"
assert_raises "exec_on $HOST2 c2 $PING $C1_NAME"

docker_on $HOST1 rm -f c1
docker_on $HOST2 rm -f c2

end_suite
