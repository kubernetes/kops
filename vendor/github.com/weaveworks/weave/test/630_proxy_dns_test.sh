#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.0.78
C2=10.2.0.34
C1_NAME=c1.weave.local
C2_NAME=seetwo.weave.local
STATIC=static.name
STATIC_IP=10.9.9.9

do_assert_resolution() {
  proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C2/24 -dt --name=c2 --add-host $STATIC:$STATIC_IP -h $C2_NAME
  proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C1/24 -dt --name=c1 --add-host $STATIC:$STATIC_IP
  $1 $HOST1 c1 $C2_NAME $C2
  assert_dns_record $HOST1 c1 $STATIC $STATIC_IP
  $1 $HOST1 c2 $C1_NAME $C1
  assert_dns_record $HOST1 c2 $STATIC $STATIC_IP
  rm_containers $HOST1 c1 c2
}

start_suite "Proxy registers containers with dns"

bridge_ip=$(weave_on $HOST1 docker-bridge-ip)

# Assert behaviour without weaveDNS
weave_on $HOST1 launch-proxy
do_assert_resolution assert_no_dns_record

# Assert behaviour with weaveDNS running
weave_on $HOST1 launch-router
do_assert_resolution assert_dns_record
weave_on $HOST1 stop-proxy

# Assert behaviour with weaveDNS running, but dns forced off
weave_on $HOST1 launch-proxy --without-dns
do_assert_resolution assert_no_dns_record

end_suite
