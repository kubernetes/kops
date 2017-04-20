#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.0.78
C2=10.2.0.79
C3=10.2.0.80

REVERSE_C1_LOWER=78.0.2.10.in-addr.arpa
REVERSE_C1_UPPER=78.0.2.10.IN-ADDR.ARPA

start_suite "DNS lookup case (in)sensitivity"

weave_on $HOST1 launch

start_container_with_dns $HOST1 --name=test

start_container $HOST1 $C1/24 --name=seeone
assert_dns_record $HOST1 test seeone.weave.local $C1
assert_dns_record $HOST1 test SeeOne.weave.local $C1
assert_dns_record $HOST1 test SEEONE.weave.local $C1

# Test reverse DNS using explicit in-addr.arpa format, lower and upper case
assert "exec_on $HOST1 test dig +short -t PTR $REVERSE_C1_LOWER" seeone.weave.local.
assert "exec_on $HOST1 test dig +short -t PTR $REVERSE_C1_UPPER" seeone.weave.local.

start_container $HOST1 $C2/24 --name=SeEtWo
assert_dns_record $HOST1 test seetwo.weave.local $C2
assert_dns_record $HOST1 test SeeTwo.weave.local $C2
assert_dns_record $HOST1 test SEETWO.weave.local $C2

start_container $HOST1 $C3/24 --name=seetwo
assert_dns_record $HOST1 test seetwo.weave.local $C2 $C3
assert_dns_record $HOST1 test SeeTwo.weave.local $C2 $C3
assert_dns_record $HOST1 test SEETWO.weave.local $C2 $C3
assert "exec_on $HOST1 test dig +short seetwo.weave.local A | grep -v ';;' | wc -l" 2

end_suite
