#! /bin/bash

. "$(dirname "$0")/config.sh"

entrypoint() {
  docker_on $HOST1 inspect --format="{{.Config.Entrypoint}}" "$@"
}

check_iface_ready() {
    assert_raises "proxy docker_on $HOST1 run -e 'WEAVE_CIDR=$1' $BASE_IMAGE $CHECK_ETHWE_UP"
    assert_raises "proxy docker_on $HOST1 run -e 'WEAVE_CIDR=$1' $BASE_IMAGE ip -4 -o addr show ethwe | grep -q '$1'"
}

start_suite "Proxy waits for weave to be ready before running container commands"

BASE_IMAGE=busybox
# Ensure the base image does not exist, so that it will be pulled
! docker_on $HOST1 inspect --format=" " $BASE_IMAGE >/dev/null 2>&1 || docker_on $HOST1 rmi $BASE_IMAGE

# check that interface is ready, with and without --no-multicast-route
weave_on $HOST1 launch-proxy --no-multicast-route
check_iface_ready 10.2.1.1/24
weave_on $HOST1 stop-proxy
weave_on $HOST1 launch-proxy
check_iface_ready 10.2.1.1/24

# Check committed containers only have one weavewait prepended
proxy_start_container $HOST1 --name c1 -e 'WEAVE_CIDR=10.2.1.1/24'
COMMITTED_IMAGE=$(proxy docker_on $HOST1 commit c1)
assert_raises "proxy docker_on $HOST1 run --name c2 $COMMITTED_IMAGE"
assert "entrypoint c2" "$(entrypoint $COMMITTED_IMAGE)"

# Check exec works on containers without weavewait
docker_on $HOST1 run -dit --name c3 $SMALL_IMAGE /bin/sh
assert_raises "proxy docker_on $HOST1 exec c3 true"

# Check we can't modify weavewait
assert_raises "proxy docker_on $HOST1 run -e 'WEAVE_CIDR=10.2.1.2/24' $BASE_IMAGE touch /w/w" 1

# Check only user-specified volumes and /w are mounted
dirs_sans_proxy=$(docker_on $HOST1 run --rm $SMALL_IMAGE mount | wc -l)
dirs_with_proxy=$(proxy docker_on $HOST1 run --rm -v /tmp/1:/srv1 -v /tmp/2:/srv2 -e 'WEAVE_CIDR=10.2.1.3/24' $SMALL_IMAGE mount | wc -l)
assert "echo $((dirs_with_proxy-3))" $dirs_sans_proxy

# Check errors are returned (when docker returns an error code)
assert_raises "proxy docker_on $HOST1 run -e 'WEAVE_CIDR=10.2.1.3/24' $SMALL_IMAGE foo 2>&1 | grep 'exec: \"foo\": executable file not found in \$PATH'"
# Check errors still happen when no command is specified
assert_raises "proxy docker_on $HOST1 run -e 'WEAVE_CIDR=10.2.1.3/24' $SMALL_IMAGE 2>&1 | grep 'Error response from daemon: No command specified'"

end_suite
