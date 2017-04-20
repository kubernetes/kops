#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.0.78
C2=10.2.0.34
C3=10.2.0.57
ENAME1=qiuds71y827hdi-seeone-1io9qd9i0wd
HNAME1=seeone
FQDN1=seeone.weave.local
ENAME2=124DJKSNK812-seetwo-128hbaJ881
HNAME2=seetwo
FQDN2=$HNAME2.weave.local
ENAME3=doesnotmatchpattern
HNAME3=doesnotmatchpattern
FQDN3=$HNAME3.weave.local

EXPANDED_NAMES="$ENAME1 $ENAME2 $ENAME3"
HOSTNAMES="$HNAME1 $HNAME2 $HNAME3"

check_dns_records() {
    ARG_PREFIX=$1
    shift
    CID1=$(proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C1/24 ${ARG_PREFIX}$1)
    CID2=$(proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C2/24 ${ARG_PREFIX}$2)
    CID3=$(proxy_start_container_with_dns $HOST1 -e WEAVE_CIDR=$C3/24 ${ARG_PREFIX}$3)

    assert_dns_a_record $HOST1 $CID1 $FQDN1 $C1
    assert_dns_a_record $HOST1 $CID2 $FQDN3 $C3
    assert_dns_a_record $HOST1 $CID2 $FQDN2 $C2

    rm_containers $HOST1 $CID1 $CID2 $CID3
}

test_setup() {
    weave_on $HOST1 launch-proxy $@
}

test_cleanup() {
    weave_on $HOST1 stop-proxy
}


start_suite "Hostname derivation"

weave_on $HOST1 launch-router

# Hostname derivation through container name substitutions
test_setup --hostname-match '^[^-]+-(?P<appname>[^-]*)-[^-]+$' --hostname-replacement '$appname'
# check that the normal container_name->hostname derivation doesn't break
check_dns_records --name= $HOSTNAMES
check_dns_records --name= $EXPANDED_NAMES
test_cleanup

# Hostname derivation from container labels
test_setup --hostname-from-label hostname-label
check_dns_records --name= $HOSTNAMES
check_dns_records --label=hostname-label= $HOSTNAMES
test_cleanup

# Hostname derivation combining container labels and substitutions
test_setup --hostname-from-label hostname-label --hostname-match '^[^-]+-(?P<appname>[^-]*)-[^-]+$' --hostname-replacement '$appname'
check_dns_records --name= $HOSTNAMES
check_dns_records --name= $EXPANDED_NAMES
check_dns_records --label=hostname-label= $HOSTNAMES
check_dns_records --label=hostname-label= $EXPANDED_NAMES
test_cleanup

end_suite
