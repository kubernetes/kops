#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.56.34
C2=10.2.54.91
DOMAIN=foo.bar
NAME=seetwo.$DOMAIN

start_suite "Resolve names in custom domain"

# Check with a trailing dot on domain
weave_on $HOST1 launch --dns-domain $DOMAIN.

start_container          $HOST1 $C2/24 --name=c2 -h $NAME
start_container_with_dns $HOST1 $C1/24 --name=c1

assert_dns_record $HOST1 c1 $NAME $C2

rm_containers $HOST1 c1
rm_containers $HOST1 c2
weave_on $HOST1 stop

# Check without a trailing dot on domain
weave_on $HOST1 launch --dns-domain $DOMAIN

start_container          $HOST1 $C2/24 --name=c2 -h $NAME
start_container_with_dns $HOST1 $C1/24 --name=c1

assert_dns_record $HOST1 c1 $NAME $C2

end_suite
