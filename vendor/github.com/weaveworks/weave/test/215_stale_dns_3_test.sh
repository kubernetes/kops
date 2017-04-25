#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Check propagation of DNS entries when gossip is delayed"

weave_on $HOST1 launch-router --no-discovery
weave_on $HOST2 launch-router --no-discovery $HOST1
weave_on $HOST3 launch-router --no-discovery $HOST2

start_container $HOST1 --name=c1
C1=$(container_ip $HOST1 c1)
start_container_with_dns $HOST3 --name=c3

assert_dns_record $HOST3 c3 c1.weave.local $C1

# Stall the propagation of gossip, stop weave and kill the container
run_on $HOST3 "sudo iptables -A INPUT -s $HOST2 -j DROP"

weave_on $HOST1 stop-router
docker_on $HOST1 rm -f c1

weave_on $HOST1 launch-router --no-discovery $HOST3

# Now re-enable gossip from host2 to host3
run_on $HOST3 "sudo iptables -D INPUT -s $HOST2 -j DROP"

sleep 1

assert_no_dns_record $HOST3 c3 c1.weave.local

end_suite
