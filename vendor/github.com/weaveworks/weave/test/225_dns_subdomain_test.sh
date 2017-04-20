#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.54.34
C2=10.2.54.91
C3=10.2.54.92
DOMAIN=foo.bar
SUBDOMAIN1=sub1.$DOMAIN
SUBDOMAIN2=sub2.$DOMAIN

start_suite "Resolve names in custom domain with subdomains"

weave_on $HOST1 launch --log-level=debug --dns-domain $DOMAIN.

start_container_with_dns $HOST1 $C1/24 --name=c1 -h foo.$SUBDOMAIN1
start_container_with_dns $HOST1 $C2/24 --name=c2 -h foo.$SUBDOMAIN2
start_container_with_dns $HOST1 $C3/24 --name=c3 -h tre.$SUBDOMAIN1 --dns-search=$SUBDOMAIN2

assert_dns_record   $HOST1 c1 foo.$SUBDOMAIN2 $C2
assert_dns_a_record $HOST1 c3 foo             $C2 foo.$SUBDOMAIN2
assert_dns_record   $HOST1 c2 foo.$SUBDOMAIN1 $C1
assert_dns_a_record $HOST1 c1 tre             $C3 tre.$SUBDOMAIN1

end_suite
