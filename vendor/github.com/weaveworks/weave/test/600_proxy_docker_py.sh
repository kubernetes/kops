#! /bin/bash
# Explicitly not called _test.sh - this isn't run, but imported by other tests.

. "$(dirname "$0")/config.sh"

IMAGE="joffrey/docker-py:1.9.0-rc2"

docker_py_test() {
    SHARD=$1
    TOTAL_SHARDS=$2

    start_suite "Run docker-py test suite against the proxy"

    # Work round https://github.com/docker/docker-py/issues/852
    docker_on $HOST1 pull busybox:buildroot-2014.02 >/dev/null
    # Get a list of the tests for use to shard
    docker_on $HOST1 pull $IMAGE >/dev/null
    CANDIDATES=$(docker_on $HOST1 run \
      $IMAGE \
      py.test --collect-only tests/integration/ \
      | sed -En "s/\s*<Module '([[:print:]]+)'>/\1/p")

    i=0
    TESTS=
    for test in $CANDIDATES; do
        if [ $(($i % $TOTAL_SHARDS)) -eq $SHARD ]; then
              TESTS="$TESTS $test"
        fi
        i=$(($i + 1))
    done

    weave_on $HOST1 launch-proxy --no-default-ipalloc

    DOCKER_BRIDGE_IP=$(weave_on $HOST1 docker-bridge-ip)

    if docker_on $HOST1 run \
        -e NOT_ON_HOST=true \
        -e DOCKER_HOST=tcp://$DOCKER_BRIDGE_IP:12375 \
        -v /tmp:/tmp \
        -v /var/run/docker.sock:/var/run/docker.sock \
        $IMAGE py.test $TESTS ; then
        assert_raises "true"
    else
        assert_raises "false"
    fi

    end_suite
}
