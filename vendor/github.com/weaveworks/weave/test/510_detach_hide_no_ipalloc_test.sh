#! /bin/bash

UNIVERSE=10.2.3.0/24

check() {
    CMD="weave_on $HOST1 $@"
    assert_raises "timeout 10 cat <( $CMD )"
}

. "$(dirname "$0")/config.sh"

start_suite "'detach' and 'hide' do not require IP allocation"

weave_on $HOST1 launch --ipalloc-range $UNIVERSE --ipalloc-init consensus=2

start_container $HOST1 10.2.1.1/24 --name=c1
check detach           10.2.1.1/24 c1
check detach       net:10.2.3.0/24 c1
check detach                       c1

weave_on $HOST1 expose 10.2.1.2/24
check hide             10.2.1.2/24
check hide         net:10.2.3.0/24
check hide

end_suite
