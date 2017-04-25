#! /bin/bash

. "$(dirname "$0")/config.sh"

C1=10.2.0.78
C2=10.2.0.34
NAME=seetwo.weave.local

STATIC1=10.9.9.9
STATIC1_NAME=static1.name
STATIC2=10.10.10.10
STATIC2_NAME=static2.name
STATICS_ARGS="--add-host=$STATIC1_NAME:$STATIC1 --add-host $STATIC2_NAME:$STATIC2"

start_suite "Resolve names on a single host"

weave_on $HOST1 launch

start_container          $HOST1 $C2/24 --name=c2 $STATICS_ARGS -h $NAME
start_container_with_dns $HOST1 $C1/24 --name=c1 $STATICS_ARGS

assert_dns_record $HOST1 c1 $NAME $C2

# Check that 'weave expose -h' names are added to DNS
weave_on $HOST1 expose
EXPOSE=$(weave_on $HOST1 expose -h testexpose1.weave.local)
weave_on $HOST1 expose -h testexpose2.weave.local
assert_dns_record $HOST1 c1 testexpose1.weave.local $EXPOSE
assert_dns_record $HOST1 c1 testexpose2.weave.local

# check we respect the --add-host arguments
for C in c1 c2 ; do
    assert_dns_a_record $HOST1 $C $STATIC1_NAME $STATIC1
    assert_dns_a_record $HOST1 $C $STATIC2_NAME $STATIC2
done

end_suite
