#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.0.67
C2=10.2.0.43
NAME=seetwo.weave.local

DOPTS="--name=c2 -h $NAME"

check() {
    assert_dns_record $HOST1 c1 $NAME $C2
    rm_containers $HOST1 c2
    assert_no_dns_record $HOST1 c1 $NAME
}

start_suite "Automatic DNS record removal on container death"

weave_on $HOST1 launch
start_container_with_dns $HOST1 $C1/24 --name=c1

start_container $HOST1 $C2/24 $DOPTS
check

docker_on $HOST1 run    $DOPTS -dt $SMALL_IMAGE /bin/sh
weave_on  $HOST1 attach $C2/24 c2
check

end_suite
