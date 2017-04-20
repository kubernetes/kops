#!/bin/bash

set -e

# Signal failures in lock file, in order to fail fast:
function signal_failure() {
    echo "KO" >"$TEST_VMS_PROV_AND_CONF_LOCK_FILE"
    exit 1
}
trap signal_failure ERR

source "$STATE"

function install_terraform() {
    curl -fsS https://releases.hashicorp.com/terraform/0.8.5/terraform_0.8.5_linux_amd64.zip | gunzip >terraform && chmod +x terraform && sudo mv terraform /usr/bin
}

function install_ansible() {
    sudo apt-get update || true
    sudo apt-get install -qq -y python-pip python-dev libffi-dev libssl-dev && pip install --user -U cffi && pip install --user ansible
    export PATH="$PATH:$HOME/.local/bin"
}

if [ -n "$TEST_AND_PUBLISH" ]; then
    [ -n "$SECRET_KEY" ] || {
        echo "Cannot run smoke tests: no secret key"
        exit 1
    }

    install_terraform >>"$TEST_VMS_SETUP_OUTPUT_FILE" 2>&1
    install_ansible >>"$TEST_VMS_SETUP_OUTPUT_FILE" 2>&1

    # Only attempt to create GCP image in first container, wait for it to be created otherwise:
    [ "$CIRCLE_NODE_INDEX" != "0" ] && export CREATE_IMAGE=0

    # Provision and configure testing VMs:
    cd "$SRCDIR/test" # Ensures we generate Terraform state files in the right folder, for later use by integration tests.
    ./run-integration-tests.sh configure >>"$TEST_VMS_SETUP_OUTPUT_FILE" 2>&1
    echo "OK" >"$TEST_VMS_PROV_AND_CONF_LOCK_FILE"
    echo "Test VMs now provisioned and configured. $(date)." >>"$TEST_VMS_SETUP_OUTPUT_FILE"
fi
