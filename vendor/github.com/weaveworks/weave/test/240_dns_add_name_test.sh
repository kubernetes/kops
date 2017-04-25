#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.0.78
C2=10.2.0.34
C3=10.2.0.12
NAME1=seeone.weave.local
NAME2=seetwo.weave.local
NAME3=seethree.weave.local

start_suite "Add and remove names on a single host"

weave_on $HOST1 launch

start_container          $HOST1 --without-dns $C2/24 --name=c2
start_container_with_dns $HOST1               $C1/24 --name=c1

weave_on $HOST1 dns-add $C2 c2 -h $NAME2

assert_dns_record $HOST1 c1 $NAME2 $C2

weave_on $HOST1 dns-add $C1 c1 -h $NAME1
weave_on $HOST1 dns-add c1 -h $NAME3

assert_dns_a_record $HOST1 c1 $NAME1 $C1
assert_dns_a_record $HOST1 c1 $NAME3 $C1

weave_on $HOST1 dns-remove $C1 c1 -h $NAME1

assert_no_dns_record $HOST1 c1 $NAME1
assert_dns_a_record $HOST1 c1 $NAME3 $C1

weave_on $HOST1 dns-remove c1 -h $NAME3

assert_no_dns_record $HOST1 c1 $NAME3

weave_on $HOST1 dns-add $C3 -h $NAME1
assert_dns_record $HOST1 c1 $NAME1 $C3

weave_on $HOST1 dns-remove $C3 -h $NAME1
assert_no_dns_record $HOST1 c1 $NAME1

end_suite
