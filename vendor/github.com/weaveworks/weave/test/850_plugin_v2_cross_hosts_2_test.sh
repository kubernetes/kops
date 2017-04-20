#! /bin/bash

. "$(dirname "$0")/config.sh"

PLUGIN_NAME="weave-ci-registry:5000/weaveworks/net-plugin"
HOST1_IP=$($SSH $HOST1 getent ahosts $HOST1 | grep "RAW" | cut -d' ' -f1)
SERVICE="weave-850-service"
NETWORK="weave-850-network"

setup_master() {
    # Setup Docker image registry on $HOST1
    docker_on $HOST1 run -p 5000:5000 -d registry:2

    # Build plugin-v2 on $HOST1, because Circle "CI" runs only ancient
    # version of Docker which does not support v2 plugins.
    build_plugin_img="buildpluginv2-ci"
    rsync -az -e "$SSH" "$(dirname $0)/../prog/net-plugin" $HOST1:~/
    $SSH $HOST1<<EOF
        docker rmi $build_plugin_img 2>/dev/null
        docker plugin disable $PLUGIN_NAME >/dev/null
        docker plugin remove $PLUGIN_NAME 2>/dev/null

        WORK_DIR=$(mktemp -d)
        mkdir -p \${WORK_DIR}/rootfs

        docker create --name=$build_plugin_img weaveworks/weave true
        docker export $build_plugin_img | tar -x -C \${WORK_DIR}/rootfs
        cp \${HOME}/net-plugin/launch.sh \${WORK_DIR}/rootfs/home/weave/launch.sh
        cp \${HOME}/net-plugin/config.json \${WORK_DIR}
        docker plugin create $PLUGIN_NAME \${WORK_DIR}

        echo "$HOST1_IP weave-ci-registry" | sudo tee -a /etc/hosts
        docker plugin push $PLUGIN_NAME

        # Start Swarm Manager and enable the plugin
        docker swarm init --advertise-addr=$HOST1_IP

        [ -n "$COVERAGE" ] && docker plugin set $PLUGIN_NAME EXTRA_ARGS="-test.coverprofile=/home/weave/cover.prof --"
        #docker plugin set $PLUGIN_NAME WEAVE_PASSWORD="foobar"
        docker plugin enable $PLUGIN_NAME
EOF
}

setup_worker() {
    $SSH $HOST2<<EOF
        echo "$HOST1_IP weave-ci-registry" | sudo tee -a /etc/hosts
        ping -nq -W 2 -c 1 weave-ci-registry
        docker swarm join --token "$1" "${HOST1_IP}:2377"
        docker plugin install --disable --grant-all-permissions $PLUGIN_NAME

        [ -n "$COVERAGE" ] && docker plugin set $PLUGIN_NAME EXTRA_ARGS="-test.coverprofile=/home/weave/cover.prof --"
        #docker plugin set $PLUGIN_NAME WEAVE_PASSWORD="foobar"
        docker plugin enable $PLUGIN_NAME
EOF
}

cleanup() {
    for h in $@; do
        $SSH $h<<EOF
            sudo sed -i '/weave-ci-registry/d' /etc/hosts
            docker service rm $SERVICE || true
            docker swarm leave --force || true
            docker network rm $NETWORK || true
            docker plugin disable -f $PLUGIN_NAME || true
            docker plugin remove -f $PLUGIN_NAME || true
EOF
    done
}

wait_for_network() {
    for i in $(seq 10); do
        if docker_on $1 network inspect $2 >/dev/null 2>&1; then
            return
        fi
        echo "Waiting for \"$2\" network"
        sleep 2
    done
    echo "Failed to wait for \"$2\" network" >&2
    return 1
}

wait_for_service() {
    for i in $(seq 60); do
        replicas=$($SSH $1 docker service ls -f="name=$2" | awk '{print $4}' | grep -v REPLICAS)
        if [ "$replicas" == "$3/$3" ]; then
            return
        fi
        echo "Waiting for \"$2\" service"
        sleep 2
    done
    echo "Failed to wait for \"$2\" service" >&2
    return 1
}

start_suite "Test Docker plugin-v2"

#cleanup $HOST1 $HOST2
#exit 1

setup_master
setup_worker $($SSH $HOST1 docker swarm join-token --quiet worker)

echo "Creating network and service..."

# Create network and service
$SSH $HOST1<<EOF
    docker plugin ls
    docker network create --driver="${PLUGIN_NAME}:latest" $NETWORK || true
    # Otherwise no containers will be scheduled on host2
    sleep 20
    docker service create --name=$SERVICE --network=$NETWORK --replicas=2 nginx || true
    journalctl -r -u docker.service -n30
EOF

wait_for_service $HOST1 $SERVICE 2

C1=$($SSH $HOST1 weave ps | grep -v weave:expose | awk '{print $1}')
C2_IP=$($SSH $HOST2 weave ps | grep -v weave:expose | awk '{print $3}' | cut -d/ -f1)

assert_raises "exec_on $HOST1 $C1 $PING $C2_IP"

# We do not test "weave {status,launch}", because the weave script does not detect
# plugin-v2 if its name is prefixed with a registry name.

# Failing to cleanup will make the rest of the tests to fail
cleanup $HOST1 $HOST2

end_suite
