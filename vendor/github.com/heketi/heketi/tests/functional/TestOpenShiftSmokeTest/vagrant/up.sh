#!/bin/sh


# Make sure we can install OpenShift
export ANSIBLE_TIMEOUT=60

vagrant up --no-provision $@ \
    && vagrant provision
