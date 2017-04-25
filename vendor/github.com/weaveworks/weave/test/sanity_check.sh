#! /bin/bash

. "$(dirname "$0")/config.sh"

set -e

begin=$(date +%s)
sanity_checks_files=${WEAVE_NET_SANITY_CHECKS_FILES:-"/tmp/weave_net_sanity_check_*.log"}

greenly echo "> Sanity-checking test machines: ping, check Docker, and check Weave..."
whitely echo Ping each host from the other

# We wrap ping and echo in a function as we want the below parallel for loop
# to exit early in case of a failed ping, and it cannot be done concisely,
# e.g. using a one-liner, without losing the status code for ping.
function check_ping() {
    local output=$(run_on $1 $PING $2)
    local status=$?
    echo "$output" >> "${sanity_checks_files/\*/$1}"
    return $status
}

pids=""
for host in $HOSTS; do
    cat >> "${sanity_checks_files/\*/$host}" <<EOF
# =====================================
# Host Ping Check: $host
# =====================================
EOF
    for other in $HOSTS; do
        if [ "$host" != "$other" ]; then
            check_ping "$host" "$other" &
            pids="$pids $!"
        fi
    done
done
for pid in $pids; do wait $pid; done


whitely echo Check we can reach docker

function check_docker() {
    docker_version=$(docker_on $1 version)
    docker_info=$(docker_on $1 info)
    docker_weave_version=$(docker_on $1 inspect -f {{.Created}} weaveworks/weave:${WEAVE_VERSION:-latest})
    weave_version=$(weave_on $1 version)
    cat >> "${sanity_checks_files/\*/$1}" << EOF
# =====================================
# Host Version Info: $1
# =====================================
# docker version
$docker_version

# docker info
$docker_info

# docker inspect -f {{.Created}} weaveworks/weave:<version>
$docker_weave_version

# weave version
$weave_version
EOF
}

pids=""
for host in $HOSTS; do
    check_docker $host &
    pids="$pids $!"
done
for pid in $pids; do wait $pid; done

greenly echo "> Sanity checks completed successfully in $(date -u -d @$(($(date +%s)-$begin)) +"%T")."
