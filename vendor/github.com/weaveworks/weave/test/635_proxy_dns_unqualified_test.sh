#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.0.78
C2=10.2.0.34
C3=10.2.0.57
C4=10.2.0.99
DOMAIN=weave.local
NAME=seeone.$DOMAIN

start_suite "Resolve unqualified names"

weave_on $HOST1 launch

proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C1/24 --name=c1 -h $NAME
proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C2/24 --name=c2 -h seetwo.$DOMAIN
proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C3/24 --name=c3 --dns-search=$DOMAIN
container=$(proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C4/24)

check() {
  assert "proxy exec_on $HOST1 $1 getent hosts seeone | tr -s ' '" "$C1 $NAME"
}

check c2
check c3
check "$container"

end_suite
