#! /bin/bash

. "$(dirname "$0")/config.sh"

UNIVERSE=10.2.2.0/24

assert_no_ethwe() {
  assert_raises "container_ip $1 $2" 1
  assert_raises "proxy exec_on $1 $2 ip link show | grep -v ethwe"
}

start_suite "Ping proxied containers over cross-host weave network (with IPAM)"

weave_on $HOST1 launch-router --ipalloc-range $UNIVERSE
weave_on $HOST2 launch-router --ipalloc-range $UNIVERSE $HOST1
weave_on $HOST1 launch-proxy
weave_on $HOST2 launch-proxy --no-default-ipalloc

proxy_start_container $HOST1 --name=auto
proxy_start_container $HOST1 --name=none       -e WEAVE_CIDR=none
proxy_start_container $HOST2 --name=zero       -e WEAVE_CIDR=
proxy_start_container $HOST2 --name=no-default
proxy_start_container $HOST1 --name=nodocker   --net=none
proxy_start_container $HOST1 --name=bridge     --net=bridge
proxy_start_container $HOST1 --name=host       --net=host
proxy_start_container $HOST1 --name=other      --net=container:auto

assert_raises "proxy exec_on $HOST1 auto $PING $(container_ip $HOST2 zero)"
assert_raises "proxy exec_on $HOST2 zero $PING $(container_ip $HOST1 auto)"
assert_raises "proxy exec_on $HOST2 zero $PING $(container_ip $HOST1 bridge)"
assert_raises "proxy exec_on $HOST2 zero $PING $(container_ip $HOST1 other)"
assert_raises "proxy exec_on $HOST2 zero $PING $(container_ip $HOST1 nodocker)"

assert_no_ethwe $HOST1 none
assert_no_ethwe $HOST2 no-default
assert_no_ethwe $HOST1 host

end_suite
