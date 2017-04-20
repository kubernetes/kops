#! /bin/bash

. "$(dirname "$0")/config.sh"

CMD="run -e WEAVE_CIDR=10.2.1.4/24 $SMALL_IMAGE $CHECK_ETHWE_UP"

check() {
  assert_raises "eval '$(weave_on $HOST1 env)' ; docker $CMD"
  assert_raises "docker $(weave_on $HOST1 config) $CMD"
}

start_suite "Configure the docker daemon for the proxy"

# No output when nothing running
assert "weave_on $HOST1 env" ""
assert "weave_on $HOST1 config" ""

weave_on $HOST1 launch-proxy
check

# Check we can use the weave script through the proxy
assert_raises "eval '$(weave_on $HOST1 env)' ; $WEAVE version"
assert_raises "eval '$(weave_on $HOST1 env)' ; $WEAVE ps"
assert_raises "eval '$(weave_on $HOST1 env)' ; $WEAVE launch-router"

# Check we can use weave env/config with multiple -Hs specified
weave_on $HOST1 stop
weave_on $HOST1 launch-proxy -H tcp://0.0.0.0:12375 -H unix:///var/run/weave/weave.sock
check

# Check we can use weave env/config with unix -Hs specified
weave_on $HOST1 stop
weave_on $HOST1 launch-proxy -H unix:///var/run/weave/weave.sock
assert_raises "run_on $HOST1 'eval \$(weave env) ; docker $CMD'"
assert_raises "run_on $HOST1 'docker \$(weave config) $CMD'"

end_suite
