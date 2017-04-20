#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.3.78
C2=10.2.3.34
C2a=10.2.3.35
C2b=10.2.3.36
UNIVERSE=10.2.4.0/24
NAME2=seetwo.weave.local
NAME4=seefour.weave.local

start_suite "Resolve names across hosts"

for host in $HOSTS; do
    weave_on $host launch --ipalloc-range $UNIVERSE $HOSTS
done

# Basic test
start_container          $HOST2 $C2/24 --name=c2 -h $NAME2
start_container_with_dns $HOST1 $C1/24 --name=c1

assert_dns_record $HOST1 c1 $NAME2 $C2

# 2 containers on each host, all with the same names
FOO_IPS=
BAR_IPS=
for host in $HOSTS; do
   CID=$(proxy_start_container $host --name=foo)
   FOO_IPS="$FOO_IPS $(container_ip $host foo)"

   CID=$(proxy_start_container $host --name=bar)
   BAR_IPS="$BAR_IPS $(container_ip $host bar)"
done

start_container_with_dns $HOST1 --name=baz
assert_dns_record $HOST1 baz foo.weave.local $FOO_IPS
assert_dns_record $HOST1 baz bar.weave.local $BAR_IPS

# now stop and start the containers a bunch of times; this tests
# gossip dns tombstone behaviour
for i in $(seq 5); do
   for host in $HOSTS; do
       proxy docker_on $host kill foo 1>/dev/null
       proxy docker_on $host kill bar 1>/dev/null
   done

   assert_no_dns_record $HOST1 baz foo.weave.local
   assert_no_dns_record $HOST1 baz bar.weave.local

   FOO_IPS=
   BAR_IPS=
   for host in $HOSTS; do
       proxy docker_on $host start foo 1>/dev/null
       proxy docker_on $host start bar 1>/dev/null

       FOO_IPS="$FOO_IPS $(container_ip $host foo)"
       BAR_IPS="$BAR_IPS $(container_ip $host bar)"
   done

   assert_dns_record $HOST1 baz foo.weave.local $FOO_IPS
   assert_dns_record $HOST1 baz bar.weave.local $BAR_IPS
done

# resolution for names mapped to multiple addresses
weave_on $HOST2 dns-add $C2a c2 -h $NAME2
weave_on $HOST2 dns-add $C2b c2 -h $NAME2
assert_dns_record $HOST1 c1 $NAME2 $C2 $C2a $C2b

# resolution when containers addresses come from IPAM
start_container          $HOST2 --name=c4 -h $NAME4
start_container_with_dns $HOST1 --name=c3
C4=$(container_ip $HOST2 c4)
assert_dns_record $HOST1 c3 $NAME4 $C4

end_suite
