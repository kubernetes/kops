#!/bin/bash
#
# Description:
#   This script runs all Weave Net's integration tests on the specified
#   provider (default: Google Cloud Platform).
#
# Usage:
#
#   Run all integration tests on Google Cloud Platform:
#   $ ./run-integration-tests.sh
#
#   Run all integration tests on Amazon Web Services:
#   PROVIDER=aws ./run-integration-tests.sh
#

set -e
DIR="$(dirname "$0")"
. "$DIR/../tools/provisioning/setup.sh" # Import gcp_on, do_on, and aws_on.
. "$DIR/config.sh"                      # Import greenly.

# Variables:
APP="weave-net"
# shellcheck disable=SC2034
PROJECT="weave-net-tests" # Only used when PROVIDER is gcp, by tools/provisioning/config.sh.
NAME=${NAME:-"$(whoami | sed -e 's/[\.\_]*//g' | cut -c 1-4)"}
PROVIDER=${PROVIDER:-gcp} # Provision using provided provider, or Google Cloud Platform by default.
NUM_HOSTS=${NUM_HOSTS:-3}
PLAYBOOK=${PLAYBOOK:-setup_weave-net_test.yml}
TESTS=${TESTS:-}
RUNNER_ARGS=${RUNNER_ARGS:-""}
# Dependencies' versions:
DOCKER_VERSION=${DOCKER_VERSION:-"$(grep -oP "(?<=DOCKER_VERSION=).*" "$DIR/../DEPENDENCIES")"}
KUBERNETES_VERSION=${KUBERNETES_VERSION:-"$(grep -oP "(?<=KUBERNETES_VERSION=).*" "$DIR/../DEPENDENCIES")"}
KUBERNETES_CNI_VERSION=${KUBERNETES_CNI_VERSION:-"$(grep -oP "(?<=KUBERNETES_CNI_VERSION=).*" "$DIR/../DEPENDENCIES")"}
# Google Cloud Platform image's name & usage (only used when PROVIDER is gcp):
IMAGE_NAME=${IMAGE_NAME:-"$(echo "$APP-docker$DOCKER_VERSION-k8s$KUBERNETES_VERSION-k8scni$KUBERNETES_CNI_VERSION" | sed -e 's/[\.\_]*//g')"}
DISK_NAME_PREFIX=${DISK_NAME_PREFIX:-$NAME}
USE_IMAGE=${USE_IMAGE:-1}
CREATE_IMAGE=${CREATE_IMAGE:-1}
CREATE_IMAGE_TIMEOUT_IN_SECS=${CREATE_IMAGE_TIMEOUT_IN_SECS:-600}
# Lifecycle flags:
SKIP_CONFIG=${SKIP_CONFIG:-}
SKIP_DESTROY=${SKIP_DESTROY:-}

function print_vars() {
    echo "--- Variables: Main ---"
    echo "PROVIDER=$PROVIDER"
    echo "NUM_HOSTS=$NUM_HOSTS"
    echo "PLAYBOOK=$PLAYBOOK"
    echo "TESTS=$TESTS"
    echo "SSH_OPTS=$SSH_OPTS"
    echo "RUNNER_ARGS=$RUNNER_ARGS"
    echo "--- Variables: Versions ---"
    echo "DOCKER_VERSION=$DOCKER_VERSION"
    echo "KUBERNETES_VERSION=$KUBERNETES_VERSION"
    echo "KUBERNETES_CNI_VERSION=$KUBERNETES_CNI_VERSION"
    echo "IMAGE_NAME=$IMAGE_NAME"
    echo "DISK_NAME_PREFIX=$DISK_NAME_PREFIX"
    echo "USE_IMAGE=$USE_IMAGE"
    echo "CREATE_IMAGE=$CREATE_IMAGE"
    echo "CREATE_IMAGE_TIMEOUT_IN_SECS=$CREATE_IMAGE_TIMEOUT_IN_SECS"
    echo "--- Variables: Flags ---"
    echo "SKIP_CONFIG=$SKIP_CONFIG"
    echo "SKIP_DESTROY=$SKIP_DESTROY"
}

function verify_dependencies() {
    local deps=(python terraform ansible-playbook gcloud)
    for dep in "${deps[@]}"; do
        if [ ! "$(which "$dep")" ]; then
            echo >&2 "$dep is not installed or not in PATH."
            exit 1
        fi
    done
}

# shellcheck disable=SC2155
function provision_locally() {
    export VAGRANT_CWD="$(dirname "${BASH_SOURCE[0]}")"
    case "$1" in
        on)
            vagrant up
            local status=$?

            # Set up SSH connection details: 
            local ssh_config=$(mktemp /tmp/vagrant_ssh_config_XXX)
            vagrant ssh-config >"$ssh_config"
            export SSH="ssh -F $ssh_config"
            # Extract username, SSH private key, and VMs' IP addresses:
            ssh_user="$(sed -ne 's/\ *User //p' "$ssh_config" | uniq)"
            ssh_id_file="$(sed -ne 's/\ *IdentityFile //p' "$ssh_config" | uniq)"
            ssh_hosts=$(sed -ne 's/Host //p' "$ssh_config")

            # Set up /etc/hosts files on this ("local") machine and the ("remote") testing machines, to map hostnames and IP addresses, so that:
            # - this machine communicates with the testing machines via their public IPs;
            # - testing machines communicate between themselves via their private IPs;
            # - we can simply use just the hostname in all scripts to refer to machines, and the difference between public and private IP becomes transparent.
            # N.B.: if you decide to use public IPs everywhere, note that some tests may fail (e.g. test #115).
            update_local_etc_hosts "$ssh_hosts" "$(for host in $ssh_hosts; do $SSH "$host" "cat /etc/hosts | grep $host"; done)"

            SKIP_CONFIG=1 # Vagrant directly configures virtual machines using Ansible -- see also: Vagrantfile
            return $status
            ;;
        off)
            vagrant destroy -f
            ;;
        *)
            echo >&2 "Unknown command $1. Usage: {on|off}."
            exit 1
            ;;
    esac
}

function setup_gcloud() {
    # Authenticate:
    gcloud auth activate-service-account --key-file "$GOOGLE_CREDENTIALS_FILE" 1>/dev/null
    # Set current project:
    gcloud config set project $PROJECT
}

function image_exists() {
    gcloud compute images list | grep "$PROJECT" | grep "$IMAGE_NAME"
}

function image_ready() {
    # GCP images seem to be listed before they are actually ready for use, 
    # typically failing the build with: "googleapi: Error 400: The resource is not ready".
    # We therefore consider the image to be ready once the disk of its template instance has been deleted.
    ! gcloud compute disks list | grep "$DISK_NAME_PREFIX"
}

function wait_for_image() {
    greenly echo "> Waiting for GCP image $IMAGE_NAME to be created..."
    for i in $(seq "$CREATE_IMAGE_TIMEOUT_IN_SECS"); do
        image_exists && image_ready && return 0
        if ! ((i % 60)); then echo "Waited for $i seconds and still waiting..."; fi
        sleep 1
    done
    redly echo "> Waited $CREATE_IMAGE_TIMEOUT_IN_SECS seconds for GCP image $IMAGE_NAME to be created, but image could not be found."
    exit 1
}

# shellcheck disable=SC2155
function create_image() {
    if [[ "$CREATE_IMAGE" == 1 ]]; then
        greenly echo "> Creating GCP image $IMAGE_NAME..."
        local begin_img=$(date +%s)
        local num_hosts=1
        terraform apply -input=false -var "app=$APP" -var "name=$NAME" -var "num_hosts=$num_hosts" "$DIR/../tools/provisioning/gcp"
        configure_with_ansible "$(terraform output username)" "$(terraform output public_ips)," "$(terraform output private_key_path)" $num_hosts
        local zone=$(terraform output zone)
        local name=$(terraform output instances_names)
        gcloud -q compute instances delete "$name" --keep-disks boot --zone "$zone"
        gcloud compute images create "$IMAGE_NAME" --source-disk "$name" --source-disk-zone "$zone" \
            --description "Testing image for Weave Net based on $(terraform output image), Docker $DOCKER_VERSION, Kubernetes $KUBERNETES_VERSION and Kubernetes CNI $KUBERNETES_CNI_VERSION."
        gcloud compute disks delete "$name" --zone "$zone"
        terraform destroy -force "$DIR/../tools/provisioning/gcp"
        rm terraform.tfstate*
        echo
        greenly echo "> Created GCP image $IMAGE_NAME in $(date -u -d @$(($(date +%s) - begin_img)) +"%T")."
    else
        wait_for_image
    fi
}

function use_or_create_image() {
    setup_gcloud
    image_exists || create_image
    export TF_VAR_gcp_image="$IMAGE_NAME" # Override the default image name.
    export SKIP_CONFIG=1                  # No need to configure the image, since already done when making the template
}

function update_local_etc_hosts() {
    echo "> Updating local /etc/hosts..."
    # Remove old entries (if present):
    for host in $1; do sudo sed -i "/$host/d" /etc/hosts; done
    # Add new entries:
    sudo sh -c "echo \"$2\" >> /etc/hosts"
}

function upload_etc_hosts() {
    # Remove old entries (if present):
    # shellcheck disable=SC2016,SC2086
    $SSH $3 'for host in '$1'; do sudo sed -i "/$host/d" /etc/hosts; done'
    # Add new entries:
    echo "$2" | $SSH "$3" "sudo -- sh -c \"cat >> /etc/hosts\""
}

function update_remote_etc_hosts() {
    echo "> Updating remote /etc/hosts..."
    local pids=""
    for host in $1; do
        upload_etc_hosts "$1" "$2" "$host" &
        local pids="$pids $!"
    done
    for pid in $pids; do wait "$pid"; done
}

# shellcheck disable=SC2155
function set_hosts() {
    export HOSTS="$(echo "$ssh_hosts" | tr '\n' ' ')"
}

function provision_remotely() {
    case "$1" in
        on)
            terraform apply -input=false -parallelism="$NUM_HOSTS" -var "app=$APP" -var "name=$NAME" -var "num_hosts=$NUM_HOSTS" "$DIR/../tools/provisioning/$2"
            local status=$?
            ssh_user=$(terraform output username)
            ssh_id_file=$(terraform output private_key_path)
            ssh_hosts=$(terraform output hostnames)
            export SSH="ssh -l $ssh_user -i $ssh_id_file $SSH_OPTS"

            # Set up /etc/hosts files on this ("local") machine and the ("remote") testing machines, to map hostnames and IP addresses, so that:
            # - this machine communicates with the testing machines via their public IPs;
            # - testing machines communicate between themselves via their private IPs;
            # - we can simply use just the hostname in all scripts to refer to machines, and the difference between public and private IP becomes transparent.
            # N.B.: if you decide to use public IPs everywhere, note that some tests may fail (e.g. test #115).
            update_local_etc_hosts "$ssh_hosts" "$(terraform output public_etc_hosts)"
            update_remote_etc_hosts "$ssh_hosts" "$(terraform output private_etc_hosts)"

            return $status
            ;;
        off)
            terraform destroy -force "$DIR/../tools/provisioning/$2"
            ;;
        *)
            echo >&2 "Unknown command $1. Usage: {on|off}."
            exit 1
            ;;
    esac
}

# shellcheck disable=SC2155
function provision() {
    local action=$([ "$1" == "on" ] && echo "Provisioning" || echo "Shutting down")
    echo
    greenly echo "> $action test host(s) on [$PROVIDER]..."
    local begin_prov=$(date +%s)
    case "$2" in
        'aws')
            aws_on
            provision_remotely "$1" "$2"
            ;;
        'do')
            do_on
            provision_remotely "$1" "$2"
            ;;
        'gcp')
            gcp_on
            [[ "$1" == "on" ]] && [[ "$USE_IMAGE" == 1 ]] && use_or_create_image
            provision_remotely "$1" "$2"
            ;;
        'vagrant')
            provision_locally "$1"
            ;;
        *)
            echo >&2 "Unknown provider $2. Usage: PROVIDER={gcp|aws|do|vagrant}."
            exit 1
            ;;
    esac
    [ "$1" == "on" ] && set_hosts
    echo
    greenly echo "> Provisioning took $(date -u -d @$(($(date +%s) - begin_prov)) +"%T")."
}

function configure_with_ansible() {
    ansible-playbook -u "$1" -i "$2" --private-key="$3" --forks="${4:-$NUM_HOSTS}" \
        --ssh-extra-args="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null" \
        --extra-vars "docker_version=$DOCKER_VERSION kubernetes_version=$KUBERNETES_VERSION kubernetes_cni_version=$KUBERNETES_CNI_VERSION" \
        "$DIR/../tools/config_management/$PLAYBOOK"
}

# shellcheck disable=SC2155
function configure() {
    echo
    if [ -n "$SKIP_CONFIG" ]; then
        greenly echo "> Skipped configuration of test host(s)."
    else
        greenly echo "> Configuring test host(s)..."
        local begin_conf=$(date +%s)
        local inventory_file=$(mktemp /tmp/ansible_inventory_XXXXX)
        echo "[all]" >"$inventory_file"
        # shellcheck disable=SC2001
        echo "$2" | sed "s/$/:$3/" >>"$inventory_file"

        # Configure the provisioned machines using Ansible, allowing up to 3 retries upon failure (e.g. APT connectivity issues, etc.):
        for i in $(seq 3); do
            configure_with_ansible "$1" "$inventory_file" "$4" && break || echo >&2 "#$i: Ansible failed. Retrying now..."
        done

        echo
        greenly echo "> Configuration took $(date -u -d @$(($(date +%s) - begin_conf)) +"%T")."
    fi
}

# shellcheck disable=SC2155
function run_tests() {
    echo
    greenly echo "> Running tests..."
    local begin_tests=$(date +%s)
    set +e # Do not fail this script upon test failure, since we need to shut down the test cluster regardless of success or failure.
    "$DIR/run_all.sh" "$@"
    local status=$?
    echo
    greenly echo "> Tests took $(date -u -d @$(($(date +%s) - begin_tests)) +"%T")."
    return $status
}

function end() {
    echo
    echo "> Build took $(date -u -d @$(($(date +%s) - begin)) +"%T")."
}

function echo_export_hosts() {
    exec 1>&111
    # Print a command to set HOSTS in the calling script, so that subsequent calls to
    # test scripts can point to the right testing machines while developing:
    echo "export HOSTS=\"$HOSTS\""
    exec 1>&2
}

function main() {
    # Keep a reference to stdout in another file descriptor (FD #111), and then globally redirect all stdout to stderr.
    # This is so that HOSTS can be eval'ed in the calling script using:
    #   $ eval $(./run-integration-tests.sh [provision|configure|setup])
    # and ultimately subsequent calls to test scripts can point to the right testing machines during development.
    if [ "$1" == "provision" ] || [ "$1" == "configure" ] || [ "$1" == "setup" ]; then
        exec 111>&1 # 111 ought to match the file descriptor used in echo_export_hosts.
        exec 1>&2
    fi

    begin=$(date +%s)
    trap end EXIT

    print_vars
    verify_dependencies

    case "$1" in
        "") # Provision, configure, run tests, and destroy test environment:
            provision on "$PROVIDER"
            configure "$ssh_user" "$ssh_hosts" "${ssh_port:-22}" "$ssh_id_file"
            "$DIR/setup.sh"
            run_tests "$TESTS"
            status=$?
            [ -z "$SKIP_DESTROY" ] && provision off "$PROVIDER"
            exit $status
            ;;

        provision)
            provision on "$PROVIDER"
            echo_export_hosts
            ;;

        configure)
            provision on "$PROVIDER" # Vagrant and Terraform do not provision twice if VMs are already provisioned, so we just set environment variables.
            configure "$ssh_user" "$ssh_hosts" "${ssh_port:-22}" "$ssh_id_file"
            echo_export_hosts
            ;;

        setup)
            provision on "$PROVIDER" # Vagrant and Terraform do not provision twice if VMs are already provisioned, so we just set environment variables.
            "$DIR/setup.sh"
            echo_export_hosts
            ;;

        test)
            provision on "$PROVIDER" # Vagrant and Terraform do not provision twice if VMs are already provisioned, so we just set environment variables.
            run_tests "$TESTS"
            ;;

        destroy)
            provision off "$PROVIDER"
            ;;

        *)
            echo "Unknown command: $1" >&2
            exit 1
            ;;
    esac
}

main "$@"
