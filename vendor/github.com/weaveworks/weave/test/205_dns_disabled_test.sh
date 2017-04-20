#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.0.87

start_suite "Standard behaviour when weaveDNS is disabled"

weave_on $HOST1 launch --no-dns

assert "weave_on $HOST1 dns-args" ""

C=$(start_container_with_dns $HOST1 $C1/24 --name=c1)

# external name resolution works
assert_raises "exec_on $HOST1 c1 getent hosts www.weave.works"

end_suite
