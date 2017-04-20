#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Negative DNS queries"

weave_on $HOST1 launch
start_container_with_dns $HOST1 --name c1

# unsupported query types, unknown names, and unknown domains
assert_raises "exec_on $HOST1 c1 dig MX c1.weave.local | grep -q 'status: NOERROR'"
assert_raises "exec_on $HOST1 c1 dig A  xx.weave.local | grep -q 'status: NXDOMAIN'"
assert_raises "exec_on $HOST1 c1 dig A  xx.invalid     | grep -q 'status: NXDOMAIN'"

end_suite
