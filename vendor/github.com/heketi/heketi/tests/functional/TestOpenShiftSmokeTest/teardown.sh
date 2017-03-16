#!/bin/sh

CURRENT_DIR=`pwd`
TOP=../../..
HEKETI_SERVER_BUILD_DIR=$TOP
FUNCTIONAL_DIR=${CURRENT_DIR}/..
HEKETI_SERVER=${FUNCTIONAL_DIR}/heketi-server

source ${FUNCTIONAL_DIR}/lib.sh

teardown_vagrant
rm -rf vagrant/roles/client/files/heketi-cli \
    vagrant/roles/cluster/files/*.img \
    vagrant/templates > /dev/null 2>&1

