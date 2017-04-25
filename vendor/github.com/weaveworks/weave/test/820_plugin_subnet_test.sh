#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Test Docker Network plugin with specified subnet"

weave_on $HOST1 launch

# using ssh rather than docker -H because CircleCI docker client is older
$SSH $HOST1 docker network create --driver weavemesh --ipam-driver weavemesh --subnet 10.40.0.0/16 testsubnet

$SSH $HOST1 docker run --name=c1 -dt --net=testsubnet --ip 10.40.0.1 $SMALL_IMAGE /bin/sh
$SSH $HOST1 docker run --name=c2 -dt --net=testsubnet                $SMALL_IMAGE /bin/sh

assert "container_ip $HOST1 c1" "10.40.0.1"
assert "container_ip $HOST1 c2" "10.40.0.2" # assuming linear allocation strategy
assert_raises "exec_on $HOST1 c1 $PING c2"
assert_raises "exec_on $HOST1 c2 $PING c1"

$SSH $HOST1 docker rm -f c1 c2
$SSH $HOST1 docker network rm testsubnet

end_suite
