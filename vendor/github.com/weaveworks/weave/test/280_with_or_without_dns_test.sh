#! /bin/bash

. "$(dirname "$0")/config.sh"

IP=10.2.0.34
TARGET=seetwo.weave.local
TARGET_IP=10.2.0.78
STATIC=static.name
STATIC_IP=10.9.9.9

check_dns() {
    chk=$1
    shift

    container=$(docker_on $HOST1 run $(dns_args $HOST1 "$@") -dt $DNS_IMAGE /bin/sh)
    weave_on $HOST1 attach $IP/24 "$@" --rewrite-hosts --add-host=$STATIC:$STATIC_IP $container
    $chk $HOST1 $container $TARGET
    assert_dns_record $HOST1 $container $STATIC $STATIC_IP
    rm_containers $HOST1 $container
}

start_suite "With or without DNS test"

# Assert behaviour without weaveDNS running
weave_on $HOST1 launch-router

start_container $HOST1 $TARGET_IP/24 --name c2 -h $TARGET

check_dns assert_no_dns_record --without-dns
check_dns assert_dns_record

end_suite
