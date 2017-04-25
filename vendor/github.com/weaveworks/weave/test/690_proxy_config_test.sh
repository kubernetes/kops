#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Various launch-proxy configurations"

# Booting it over unix socket listens on unix socket
run_on $HOST1 COVERAGE=$COVERAGE weave launch-proxy
assert_raises "run_on $HOST1 sudo docker -H unix:///var/run/weave/weave.sock ps"
assert_raises "proxy docker_on $HOST1 ps" 1
weave_on $HOST1 stop-proxy

# Booting it over tcp listens on tcp
weave_on $HOST1 launch-proxy
assert_raises "run_on $HOST1 sudo docker -H unix:///var/run/weave/weave.sock ps" 1
assert_raises "proxy docker_on $HOST1 ps"
weave_on $HOST1 stop-proxy

# Booting it over tcp (no prefix) listens on tcp
DOCKER_HOST=tcp://$HOST1:$DOCKER_PORT $WEAVE launch-proxy
assert_raises "run_on $HOST1 sudo docker -H unix:///var/run/weave/weave.sock ps" 1
assert_raises "proxy docker_on $HOST1 ps"
weave_on $HOST1 stop-proxy

# Booting it with -H outside /var/run/weave, still works
socket="$($SSH $HOST1 mktemp -d)/weave.sock"
weave_on $HOST1 launch-proxy -H unix://$socket
assert_raises "run_on $HOST1 sudo docker -H unix:///$socket ps" 0
weave_on $HOST1 stop-proxy

# Booting it against non-standard docker unix sock
run_on $HOST1 "DOCKER_HOST=unix:///var/run/alt-docker.sock COVERAGE=$COVERAGE weave launch-proxy -H tcp://0.0.0.0:12375"
assert_raises "proxy docker_on $HOST1 ps"
weave_on $HOST1 stop-proxy

# Booting it over tls errors
assert_raises "! DOCKER_CLIENT_ARGS='--tls' weave_on $HOST1 launch-proxy"
assert_raises "! DOCKER_CERT_PATH='./tls' DOCKER_TLS_VERIFY=1 weave_on $HOST1 launch-proxy"

# Booting it with a specific -H overrides defaults
weave_on $HOST1 launch-proxy -H tcp://0.0.0.0:12345
assert_raises "run_on $HOST1 sudo docker -H tcp://$HOST1:12345 ps"
assert_raises "proxy docker_on $HOST1 ps" 1
weave_on $HOST1 stop-proxy

end_suite
