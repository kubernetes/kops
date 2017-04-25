#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Check that docker inspect returns the weave IP"

weave_on $HOST1 launch-router
weave_on $HOST1 launch-proxy --rewrite-inspect

proxy docker_on $HOST1 run -dt --name c1 $SMALL_IMAGE /bin/sh
inspect_format="{{.Name}} {{.NetworkSettings.MacAddress}} {{.NetworkSettings.IPAddress}}/{{.NetworkSettings.IPPrefixLen}}"
expected="/$(weave_on $HOST1 ps c1)"

assert "proxy docker_on $HOST1 inspect --format='$inspect_format' c1" "$expected"

end_suite
