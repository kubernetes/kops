#! /bin/bash

. "$(dirname "$0")/config.sh"

IP=10.2.0.67

start_suite "Reverse DNS test"

weave_on $HOST1 launch
start_container_with_dns $HOST1 --name=c1

# We try names in alphabetical order; this is because the
# nameserver keeps entries sorted by hostname then address,
# so by starting with acontainer then trying bcontainer,
# we're guaranteed to have to ignore a tombstone to get the
# right result.
for name in acontainer bcontainer ccontainer; do
    start_container $HOST1 $IP/24 --name=$name
    assert_dns_record $HOST1 c1 $name.weave.local $IP
    rm_containers $HOST1 $name
done

end_suite
