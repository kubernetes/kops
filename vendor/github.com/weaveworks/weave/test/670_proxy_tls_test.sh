#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Boot the proxy with TLS-enabled Docker support"

PWD=$($SSH $HOST1 pwd)

weave_on $HOST1 launch-router
weave_on $HOST1 launch-proxy \
  --tlsverify \
  --tlscacert $PWD/tls/ca.pem \
  --tlscert   $PWD/tls/$HOST1.pem \
  --tlskey    $PWD/tls/$HOST1-key.pem

assert_raises "DOCKER_CERT_PATH=$(dirname "$0")/tls proxy docker_on $HOST1 --tlsverify ps"

end_suite
