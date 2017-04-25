# NB only to be sourced

set -e

DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Protect against being sourced multiple times to prevent
# overwriting assert.sh global state
if ! [ -z "$SOURCED_CONFIG_SH" ]; then
    return
fi
SOURCED_CONFIG_SH=true

# these ought to match what is in Vagrantfile
N_MACHINES=${N_MACHINES:-3}
IP_PREFIX=${IP_PREFIX:-192.168.48}
IP_SUFFIX_BASE=${IP_SUFFIX_BASE:-10}

if [ -z "$HOSTS" ] ; then
    for i in $(seq 1 $N_MACHINES); do
        IP="${IP_PREFIX}.$((${IP_SUFFIX_BASE}+$i))"
        HOSTS="$HOSTS $IP"
    done
fi

# these are used by the tests
HOST1=$(echo $HOSTS | cut -f 1 -d ' ')
HOST2=$(echo $HOSTS | cut -f 2 -d ' ')
HOST3=$(echo $HOSTS | cut -f 3 -d ' ')

. "$DIR/assert.sh"

SSH_DIR=${SSH_DIR:-$DIR}
SSH_OPTS=${SSH_OPTS:-"-o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -o PasswordAuthentication=no -o IdentitiesOnly=yes -o LogLevel=ERROR"}
SSH=${SSH:-ssh -l vagrant -i "$SSH_DIR/insecure_private_key" $SSH_OPTS}

SMALL_IMAGE="alpine"
DNS_IMAGE="aanand/docker-dnsutils"
TEST_IMAGES="$SMALL_IMAGE $DNS_IMAGE"

PING="ping -nq -W 2 -c 1"
CHECK_ETHWE_UP="grep ^1$ /sys/class/net/ethwe/carrier"
CHECK_ETHWE_MISSING="test ! -d /sys/class/net/ethwe"

DOCKER_PORT=2375

CHECKPOINT_DISABLE=true

# The regexp here is far from precise, but good enough. (taken from weave script)
IP_REGEXP="[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}"
CIDR_REGEXP="(ip:|net:)?$IP_REGEXP/[0-9]{1,2}"

is_cidr() {
    [ $1 = "net:default" ] && return
    echo "$1" | grep -E "^$CIDR_REGEXP$" >/dev/null
}

upload_executable() {
    host=$1
    file=$2
    target=${3:-/usr/local/bin/$(basename "$file")}
    dir=$(dirname "$target")
    run_on $host "[ -e '$dir' ] || sudo mkdir -p '$dir'"
    [ -z "$DEBUG" ] || greyly echo "Uploading to $host: $file -> $target" >&2
    <"$file" remote $host $SSH $host sh -c "cat | sudo tee $target >/dev/null"
    run_on $host "sudo chmod a+x $target"
}

remote() {
    rem=$1
    shift 1
    "$@" > >(while read line; do echo -e $'\e[0;34m'"$rem>"$'\e[0m'" $line"; done)
}

colourise() {
    [ -t 0 ] && echo -ne $'\e['$1'm' || true
    shift
    # It's important that we don't do this in a subshell, as some
    # commands we execute need to modify global state
    "$@"
    [ -t 0 ] && echo -ne $'\e[0m' || true
}

whitely() {
    colourise '1;37' "$@"
}

greyly () {
    colourise '0;37' "$@"
}

redly() {
    colourise '1;31' "$@"
}

greenly() {
    colourise '1;32' "$@"
}

run_on() {
    host=$1
    shift 1
    [ -z "$DEBUG" ] || greyly echo "Running on $host: $@" >&2
    remote $host $SSH $host "$@"
}

docker_on() {
    host=$1
    shift 1
    [ -z "$DEBUG" ] || greyly echo "Docker on $host:$DOCKER_PORT: $@" >&2
    docker -H tcp://$host:$DOCKER_PORT "$@"
}

docker_api_on() {
    host=$1
    method=$2
    url=$3
    data=$4
    shift 4
    [ -z "$DEBUG" ] || greyly echo "Docker (API) on $host:$DOCKER_PORT: $method $url" >&2
    echo -n "$data" | curl -s -f -X "$method" -H Content-Type:application/json "http://$host:$DOCKER_PORT/v1.15$url" -d @-
}

proxy() {
    DOCKER_PORT=12375 "$@"
}

weave_on() {
    host=$1
    shift 1
    [ -z "$DEBUG" ] || greyly echo "Weave on $host:$DOCKER_PORT: $@" >&2
    CHECKPOINT_DISABLE="$CHECKPOINT_DISABLE" DOCKER_HOST=tcp://$host:$DOCKER_PORT $WEAVE "$@"
}

stop_weave_on() {
    host=$1
    weave_on $host stop 1>/dev/null 2>&1 || true
    if [ -n "$COVERAGE" ]; then
        for C in weaveplugin weaveproxy weave ; do
            collect_coverage $host $C
        done
    fi
}

exec_on() {
    host=$1
    container=$2
    shift 2
    docker -H tcp://$host:$DOCKER_PORT exec $container "$@"
}

# Look through 'docker run' args and try to make the hostname match the name
dns_args() {
    local host=$1 NAME_ARG="" HOSTNAME_SPECIFIED=""
    shift
    weave_on $host dns-args "$@"
    while [ $# -gt 0 ] ; do
        case "$1" in
            --name)
                NAME_ARG="$2"
                shift
                ;;
            --name=*)
                NAME_ARG="${1#*=}"
                ;;
            -h|--hostname|--hostname=*)
                HOSTNAME_SPECIFIED=1
                ;;
        esac
        shift
    done
    if [ -n "$NAME_ARG" -a -z "$HOSTNAME_SPECIFIED" ] ; then
        echo " --hostname=$NAME_ARG.weave.local"
    fi
}

start_container_image() {
    local image=$1 host=$2 cidr="" weave_dns_args="" container
    shift 2
    if [ "$1" = "--without-dns" ] ; then
        shift
    else
        weave_dns_args=$(dns_args $host "$@")
    fi
    is_cidr $1 && { cidr=$1; shift; }
    container=$(docker_on $host run $weave_dns_args $name_args "$@" -dt $image /bin/sh)
    if ! weave_on $host attach $cidr $container >/dev/null ; then
        docker_on $host rm -f $container
        return 1
    fi
    echo $container
}

start_container() {
    start_container_image $SMALL_IMAGE "$@"
}

start_container_with_dns() {
    start_container_image $DNS_IMAGE "$@"
}

start_container_local_plugin() {
    host=$1
    shift 1
    # using ssh rather than docker -H because CircleCI docker client is older
    $SSH $host docker run "$@" -dt --net=weave $SMALL_IMAGE /bin/sh
}

proxy_start_container() {
    host=$1
    shift 1
    proxy docker_on $host run "$@" -dt $SMALL_IMAGE /bin/sh
}

proxy_start_container_with_dns() {
    host=$1
    shift 1
    proxy docker_on $host run "$@" -dt $DNS_IMAGE /bin/sh
}

wait_for_proxy() {
    for i in $(seq 1 120); do
        echo "Waiting for proxy to start"
        if proxy docker_on $1 info > /dev/null 2>&1 ; then
            return
        fi
        sleep 1
    done
    echo "Timed out waiting for proxy to start" >&2
    exit 1
}

rm_containers() {
    host=$1
    shift
    [ $# -eq 0 ] || docker_on $host rm -f -v "$@" >/dev/null
}

container_ip() {
    weave_on $1 ps $2 | grep -o -E '[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}'
}

container_pid() {
    docker_on $1 inspect -f '{{.State.Pid}}' $2
}

wait_for_attached() {
    host=$1
    container=$2
    for i in $(seq 1 10); do
        echo "Waiting for $container on $host to be attached"
        if exec_on $host $container $CHECK_ETHWE_UP > /dev/null 2>&1 ; then
            return
        fi
        sleep 1
    done
    echo "Timed out waiting for $container on $host to be attached" >&2
    exit 1
}

# assert_dns_record <host> <container> <name> [<ip> ...]
assert_dns_record() {
    local host=$1
    local container=$2
    local name=$3
    shift 3
    exp_ips_regex=$(echo "$@" | sed -e 's/ /\\\|/g')

    [ -z "$DEBUG" ] || greyly echo "Checking whether $name exists at $host:$container"
    assert_raises "exec_on $host $container getent hosts $name | grep -q '$exp_ips_regex'"

    [ -z "$DEBUG" ] || greyly echo "Checking whether the IPs '$@' exists at $host:$container"
    for ip in "$@" ; do
        assert "exec_on $host $container getent hosts $ip | tr -s ' ' | tr '[:upper:]' '[:lower:]'" "$(echo $ip $name | tr '[:upper:]' '[:lower:]')"
    done
}

# assert_no_dns_record <host> <container> <name>
assert_no_dns_record() {
    host=$1
    container=$2
    name=$3

    [ -z "$DEBUG" ] || greyly echo "Checking if '$name' does not exist at $host:$container"
    assert_raises "exec_on $host $container getent hosts $name" 2
}

# assert_dns_a_record <host> <container> <name> <ip> [<expected_name>]
assert_dns_a_record() {
    exp_name=${5:-$3}
    assert "exec_on $1 $2 getent hosts $3 | tr -s ' ' | cut -d ' ' -f 1,2" "$4 $exp_name"
}

# assert_dns_ptr_record <host> <container> <name> <ip>
assert_dns_ptr_record() {
    assert "exec_on $1 $2 getent hosts $4 | tr -s ' '" "$4 $3"
}

# Kill a container process and make sure it's restarted by Docker
check_restart() {
    OLD_PID=$(container_pid $1 $2)

    run_on $1 sudo kill $OLD_PID

    for i in $(seq 1 10); do
        NEW_PID=$(container_pid $1 $2)

        if [ $NEW_PID != 0 -a $NEW_PID != $OLD_PID ] ; then
            return 0
        fi

        sleep 1
    done

    return 1
}

start_suite() {
    for host in $HOSTS; do
        [ -z "$DEBUG" ] || echo "Cleaning up on $host: removing all containers and resetting weave"
        PLUGIN_FILTER=$(docker_on $host inspect -f 'grep -v {{printf "%.12s" .Id}}' weaveplugin 2>/dev/null) || PLUGIN_FILTER=cat
        rm_containers $host $(docker_on $host ps -aq 2>/dev/null | $PLUGIN_FILTER)
        weave_on $host reset 2>/dev/null
        run_on $host sudo rm -f /opt/cni/bin/weave-plugin-latest /opt/cni/bin/weave-net /opt/cni/bin/weave-ipam /etc/cni/net.d/10-weave.conf
    done
    whitely echo "$@"
}

# Common postconditions to assert on each host, after each test:
assert_common_postconditions() {
    # Ensure we do not generate any defunct (a.k.a. zombie) process:
    assert "run_on $1 ps aux | grep -c '[d]efunct'" "0"
}

end_suite() {
    for host in $HOSTS; do
        assert_common_postconditions "$host"
    done
    whitely assert_end
    for host in $HOSTS; do
        stop_weave_on $host
    done
}

collect_coverage() {
    host=$1
    container=$2
    mkdir -p ./coverage
    rm -f cover.prof
    docker_on $host cp $container:/home/weave/cover.prof . 2>/dev/null || return 0
    # ideally we'd know the name of the test here, and put that in the filename
    mv cover.prof $(mktemp -u ./coverage/integration.XXXXXXXX) || true
}

WEAVE=$DIR/../weave
DOCKER_NS=$DIR/../bin/docker-ns
