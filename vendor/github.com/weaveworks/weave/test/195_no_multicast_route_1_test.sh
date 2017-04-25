#! /bin/bash

. "$(dirname "$0")/config.sh"

show_multicast_route_on() {
    exec_on $1 $2 ip route show 224.0.0.0/4
}

start_suite "--no-multicast-route operation"

# In case one is left around from a previous test
$SSH $HOST1 docker network rm testmcasttrue testmcastfalse >/dev/null 2>&1 || true

# Ensure containers run either way have no multicast route
weave_on $HOST1 launch-router --plugin
weave_on $HOST1 launch-proxy --no-multicast-route

docker_on $HOST1 run --name c1 -dt $SMALL_IMAGE /bin/sh
weave_on $HOST1 attach --no-multicast-route c1
proxy_start_container $HOST1 --name c2
start_container_local_plugin $HOST1 --name=c3

assert "show_multicast_route_on $HOST1 c1"
assert "show_multicast_route_on $HOST1 c2"
assert "show_multicast_route_on $HOST1 c3" "224.0.0.0/4 dev ethwe0 "

# Now try via docker network options
# using ssh rather than docker -H because CircleCI docker client is older
$SSH $HOST1 docker network create --driver weavemesh --ipam-driver weavemesh --opt works.weave.multicast=false testmcastfalse
$SSH $HOST1 docker run --net=testmcastfalse --name c4 -di $SMALL_IMAGE /bin/sh
assert "show_multicast_route_on $HOST1 c4"
$SSH $HOST1 docker network create --driver weavemesh --ipam-driver weavemesh --opt works.weave.multicast testmcasttrue
$SSH $HOST1 docker run --net=testmcasttrue --name c5 -di $SMALL_IMAGE /bin/sh
assert "show_multicast_route_on $HOST1 c5" "224.0.0.0/4 dev ethwe0 "

# Ensure current proxy options are obeyed on container start
docker_on $HOST1 stop -t 1 c2
weave_on $HOST1 stop-proxy
weave_on $HOST1 launch-proxy

proxy docker_on $HOST1 start c2

assert "show_multicast_route_on $HOST1 c2" "224.0.0.0/4 dev ethwe "

docker_on $HOST1 rm -f c3 c4 c5
$SSH $HOST1 docker network rm testmcasttrue testmcastfalse

end_suite
