#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Resolve a non-weave address"

weave_on $HOST1 launch-router

start_container_with_dns $HOST1 10.2.1.5/24 --name=c1

assert_raises "exec_on $HOST1 c1 host -t mx weave.works | grep google"
assert_raises "exec_on $HOST1 c1 getent hosts 8.8.8.8   | grep google"

end_suite
