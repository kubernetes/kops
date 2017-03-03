#!/bin/sh

CURRENT_DIR=`pwd`
TOP=../../..
HEKETI_SERVER_BUILD_DIR=$TOP
FUNCTIONAL_DIR=${CURRENT_DIR}/..
HEKETI_SERVER=${FUNCTIONAL_DIR}/heketi-server
HEKETI_DOCKER_IMG=heketi-docker-ci.img
GLUSTERFS_DOCKER_IMG=gluster-docker-ci.img
DOCKERDIR=$TOP/extras/docker
CLIENTDIR=$TOP/client/cli/go

source ${FUNCTIONAL_DIR}/lib.sh

build_docker_file(){
    echo "Create Heketi Docker image"
    vagrant_heketi_docker=$CURRENT_DIR/vagrant/roles/cluster/files/$HEKETI_DOCKER_IMG
    mkdir -p vagrant/roles/cluster/files
    if [ ! -f "$vagrant_heketi_docker" ] ; then
        cd $DOCKERDIR/ci
        cp $TOP/heketi $DOCKERDIR/ci || fail "Unable to copy $TOP/heketi to $DOCKERDIR/ci"
        _sudo docker build --rm --tag heketi/heketi:ci . || fail "Unable to create docker container"
        _sudo docker save -o $HEKETI_DOCKER_IMG heketi/heketi:ci || fail "Unable to save docker image"
        cp $HEKETI_DOCKER_IMG $vagrant_heketi_docker
        _sudo docker rmi heketi/heketi:ci
        cd $CURRENT_DIR
    fi

    echo "Create GlusterFS Docker image"
    vagrant_gluster_docker=$CURRENT_DIR/vagrant/roles/cluster/files/$GLUSTERFS_DOCKER_IMG
    if [ ! -f "$vagrant_gluster_docker" ] ; then
        cd $DOCKERDIR/gluster
        _sudo docker build --rm --tag heketi/gluster:ci . || fail "Unable to create docker container"
        _sudo docker save -o $GLUSTERFS_DOCKER_IMG heketi/gluster:ci || fail "Unable to save docker image"
        cp $GLUSTERFS_DOCKER_IMG $vagrant_gluster_docker
        cd $CURRENT_DIR
    fi

}

build_heketi() {
    cd $TOP
    make || fail  "Unable to build heketi"
    cd $CURRENT_DIR
}

copy_client_files() {
    cp $CLIENTDIR/heketi-cli vagrant/roles/client/files
    cp -r $TOP/extras/openshift/templates vagrant
}

deploy_heketi_glusterfs() {
    cd tests/deploy
    _sudo ./run.sh || fail "Unable to deploy"
    cd $CURRENT_DIR
}

teardown() {
    teardown_vagrant
    rm -f $vagrant_heketi_docker $vagrant_gluster_docker > /dev/null 2>&1
}

teardown
build_heketi
copy_client_files
build_docker_file
start_vagrant
deploy_heketi_glusterfs
teardown

