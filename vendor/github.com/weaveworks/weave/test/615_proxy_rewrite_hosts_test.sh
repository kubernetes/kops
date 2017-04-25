#! /bin/bash

. "$(dirname "$0")/config.sh"

expected_ip=10.2.1.1

image_ip() {
  proxy docker_on $HOST1 run -e "WEAVE_CIDR=$expected_ip/24" $1 hostname -i | cut -d' ' -f1
}

start_suite "Proxy rewrites hosts file"

# Default rewrites the host file
weave_on $HOST1 launch-proxy
assert "image_ip $SMALL_IMAGE" $expected_ip

# When container user is non-root
docker_on $HOST1 build -t non-root >/dev/null - <<- EOF
  FROM $SMALL_IMAGE
  RUN adduser -D -s /bin/sh user
  ENV HOME /home/user
  USER user
  CMD true
EOF
assert "image_ip non-root" $expected_ip

# When rewrite hosts is disabled
weave_on $HOST1 stop-proxy
weave_on $HOST1 launch-proxy --no-rewrite-hosts
assert_raises "image_ip $SMALL_IMAGE | grep -v $expected_ip"

end_suite
