#! /bin/bash

. "$(dirname "$0")/config.sh"

build_image() {
    docker_on $HOST1 build -t $1 >/dev/null - <<- EOF
  FROM $SMALL_IMAGE
  ENTRYPOINT $2
  CMD $3
EOF
}

run_container() {
    assert_raises "proxy docker_on $HOST1 run -e 'WEAVE_CIDR=10.2.1.1/24' $1"
}

start_suite "Proxy uses correct entrypoint and command with weavewait"
weave_on $HOST1 launch-proxy --no-restart

build_image check-ethwe-up '["grep"]' '["^1$", "/sys/class/net/ethwe/carrier"]'
run_container "check-ethwe-up"

build_image grep '["grep"]' ''
run_container "grep ^1$ /sys/class/net/ethwe/carrier"

build_image false '["/bin/false"]' ''
run_container "--entrypoint='grep' false ^1$ /sys/class/net/ethwe/carrier"

weave_on $HOST1 launch-router --ipalloc-range 10.2.2.0/24
# NOTE: docker-kill hangs (https://github.com/docker/docker/issues/31447), so we
# kill directly the weaveproxy process instead.
WEAVEPROXY_PID=$(container_pid $HOST1 weaveproxy)
$SSH $HOST1 "ps aux | grep weaveproxy"
run_on $HOST1 "sudo kill -9 $WEAVEPROXY_PID"
weave_on $HOST1 launch-proxy

assert_raises "proxy docker_on $HOST1 run check-ethwe-up"

create_container_json=$(cat <<-EOF
{
  "Hostname":"c1",
  "AttachStdin":false,
  "AttachStdout":true,
  "AttachStderr":true,
  "Tty":false,
  "OpenStdin":false,
  "StdinOnce":false,
  "Image":"check-ethwe-up"
}
EOF
)

# Test a minimal request to the unversioned (v1.0) docker api
CONTAINER_ID=$(curl -s -X POST --header "Content-Type: application/json" http://$HOST1:12375/containers/create -d "$create_container_json" | sed -e 's/.*"Id":"\([^"]*\)".*/\1/')
assert_raises "proxy docker_on $HOST1 start -ai $CONTAINER_ID"

end_suite
