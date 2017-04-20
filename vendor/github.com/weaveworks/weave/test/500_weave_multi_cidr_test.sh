#! /bin/bash

. "$(dirname "$0")/config.sh"

NAME=multicidr.weave.local

# assert_container_cidrs <host> <cid> [<cidr> ...]
assert_container_cidrs() {
    local HOST=$1
    local CID=$(echo $2 | cut -b 1-12)
    shift 2
    local CIDRS="$@"

    # Assert container has attached CIDRs
    if [ -z "$CIDRS" ] ; then
        assert        "weave_on $HOST ps $CID" ""
    else
        assert_raises "weave_on $HOST ps $CID | grep -E '^$CID [0-9a-f:]{17} $CIDRS$'"
    fi
}

# assert_zone_records <host> <fqdn> [<ip>|<cidr> ...]
assert_zone_records() {
    local HOST=$1
    local FQDN=$2
    shift 2

    records=$(weave_on $HOST dns-lookup $FQDN; echo "EOF") || true
    count=$(echo -n "$records" | wc -l)
    assert "echo $count" $#
    for ADDR; do
        assert_raises "echo $records | grep '${ADDR%/*}'"
    done
}

# assert_ips_and_dns <host> <cid> <fqdn> [<cidr> ...]
assert_ips_and_dns() {
    local HOST=$1
    local CID=$2
    local FQDN=$3
    shift 3

    assert_container_cidrs $HOST $CID  "$@"
    assert_zone_records    $HOST $FQDN "$@"
}

# assert_bridge_cidrs <host> <dev> <cidr> [<cidr> ...]
assert_bridge_cidrs() {
    local HOST=$1
    local DEV=$2
    shift 2
    local CIDRS="$@"

    BRIDGE_CIDRS=$($SSH $HOST ip addr show dev $DEV | grep -o 'inet .*' | cut -d ' ' -f 2)
    assert "echo $BRIDGE_CIDRS" "$CIDRS"
}

assert_equal() {
    result=$1
    shift
    expected="$@"
    assert "echo $result" "$expected"
}

start_suite "Weave attach/detach/expose/hide with multiple cidr arguments"

# also check that these commands understand all address flavours

# NOTE: in these tests, net: arguments are checked against a
# specific address, i.e. we are assuming that IPAM always returns the
# lowest available address in the subnet

weave_on $HOST1 launch-router --ipalloc-range 10.2.3.0/24

# Run container with three cidrs
CID=$(docker_on  $HOST1 run --name=multicidr -h $NAME -dt $SMALL_IMAGE /bin/sh)
weave_on         $HOST1 attach            10.2.1.1/24 ip:10.2.2.1/24 net:10.2.3.0/24 $CID
assert_ips_and_dns     $HOST1 $CID $NAME. 10.2.1.1/24    10.2.2.1/24     10.2.3.1/24

# Remove two of them
IPS=$(weave_on         $HOST1 detach                  ip:10.2.2.1/24 net:10.2.3.0/24 $CID)
assert_equal "$IPS"                                      10.2.2.1        10.2.3.1
assert_ips_and_dns     $HOST1 $CID $NAME. 10.2.1.1/24
# ...and the remaining one
IPS=$(weave_on         $HOST1 detach      10.2.1.1/24                                $CID)
assert_equal "$IPS"                       10.2.1.1
assert_ips_and_dns     $HOST1 $CID $NAME.

# Put one back
IPS=$(weave_on         $HOST1 attach      10.2.1.1/24                                $CID)
assert_equal "$IPS"                       10.2.1.1
assert_ips_and_dns     $HOST1 $CID $NAME. 10.2.1.1/24
# ...and the remaining two
IPS=$(weave_on         $HOST1 attach                  ip:10.2.2.1/24 net:10.2.3.0/24 $CID)
assert_equal "$IPS"                                      10.2.2.1        10.2.3.1
assert_ips_and_dns     $HOST1 $CID $NAME. 10.2.1.1/24    10.2.2.1/24     10.2.3.1/24

# Expose three cidrs
IPS=$(weave_on         $HOST1 expose      10.2.1.2/24 ip:10.2.2.2/24 net:10.2.3.0/24)
assert_equal "$IPS"                       10.2.1.2       10.2.2.2        10.2.3.2
assert_bridge_cidrs    $HOST1 weave       10.2.1.2/24    10.2.2.2/24     10.2.3.2/24

# Hide two of them
IPS=$(weave_on         $HOST1 hide                    ip:10.2.2.2/24 net:10.2.3.0/24)
assert_equal "$IPS"                                      10.2.2.2        10.2.3.2
assert_bridge_cidrs    $HOST1 weave       10.2.1.2/24
# ...and the remaining one
IPS=$(weave_on         $HOST1 hide        10.2.1.2/24)
assert_equal "$IPS"                       10.2.1.2
assert_bridge_cidrs    $HOST1 weave

# Now detach and run another container to check we have released IPs in IPAM
IPS=$(weave_on         $HOST1 detach                                                 $CID)
assert_equal "$IPS"                                                      10.2.3.1
CID2=$(start_container $HOST1                                        net:10.2.3.0/24)
assert_container_cidrs $HOST1 $CID2                                      10.2.3.1/24

# Error conditions: host address not network, subnet too small
assert_raises "start_container $HOST1                                net:10.2.3.2/30" 1
assert_raises "start_container $HOST1                                net:10.2.3.2/31" 1

end_suite
