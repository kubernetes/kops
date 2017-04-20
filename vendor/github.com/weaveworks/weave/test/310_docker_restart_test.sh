#! /bin/bash

. "$(dirname "$0")/config.sh"

start_suite "Containers get same IP address on restart"

weave_on $HOST1 launch-router
weave_on $HOST1 launch-proxy

# Use up first address with throwaway container
start_container $HOST1 --name=c1
# Use sigproxy+sleep to create a container that will die when Docker asks it to.
proxy docker_on $HOST1 run -di --name=c2 --restart=always -dt --entrypoint="/home/weave/sigproxy" weaveworks/weaveexec sleep 600
C2=$(container_ip $HOST1 c2)
assert_raises "[ -n $C2 ]"
wait_for_attached $HOST1 c2

docker_on $HOST1 rm -f c1

# Restart docker daemon, using different commands for systemd- and upstart-managed.
run_on $HOST1 sh -c "command -v systemctl >/dev/null && sudo systemctl restart docker || sudo service docker restart"
wait_for_proxy $HOST1
wait_for_attached $HOST1 c2
# Check same IP address was retained
assert "container_ip $HOST1 c2" "$C2"

end_suite
