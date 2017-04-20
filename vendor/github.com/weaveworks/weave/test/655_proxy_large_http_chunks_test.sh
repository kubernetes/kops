#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Check the proxy handles arbitrarily large http chunks"
# docker images returns all its data as one giant http chunk.

base=scratch
for i in `seq 0 4`; do
  docker_on $HOST1 build -t image$i >/dev/null - <<-EOF
  FROM $base
  LABEL image${i}_biglabel "$(printf %8192s)"
EOF
  base=image$i
done

# Sanity-check that it's big enough to cause issues.
assert_raises "test $(curl -s http://$HOST1:$DOCKER_PORT/v1.19/images/json?all=true | wc -c) -gt 65536"

weave_on $HOST1 launch-proxy

assert_raises "proxy docker_on $HOST1 images -a"

end_suite
