#!/bin/bash

# shellcheck disable=SC2155
wait_for() {
    local timeout=$1
    local lock_file=$2
    for i in $(seq 1 "$timeout"); do
        if [ -f "$lock_file" ]; then
            local status="$(cat "$lock_file")"
            if [ "$status" == "OK" ]; then
                echo "[$i seconds]: $lock_file found and status: $status." && return
            else
                echo "[$i seconds]: $lock_file found and status: $status." && return 1
            fi
        fi
        if ! ((i % 10)); then echo "[$i seconds]: Waiting for $lock_file to be created..."; fi
        sleep 1
    done
    echo "Timed out waiting for test VMs to be ready. See details in: $TEST_VMS_SETUP_OUTPUT_FILE" >&2
    return 1
}
