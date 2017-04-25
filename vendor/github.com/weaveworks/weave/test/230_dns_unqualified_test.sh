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

start_container          $HOST1 $C1/24 --name=c1 -h $NAME
start_container_with_dns $HOST1 $C2/24 --name=c2 -h seetwo.$DOMAIN
start_container_with_dns $HOST1 $C3/24 --name=c3 --dns-search=$DOMAIN
container=$(start_container_with_dns $HOST1 $C4/24)

assert_dns_a_record $HOST1 c2           seeone     $C1 $NAME
assert_dns_a_record $HOST1 c3           seeone     $C1 $NAME
assert_dns_a_record $HOST1 "$container" seeone     $C1 $NAME

# check that unqualified names are automatically qualified for broken resolvers
weave_on $HOST1 dns-add $C1 c1 -h mysql.weave.local
assert_dns_a_record $HOST1 c1 mysql $C1

end_suite
