#!/usr/bin/bash

CLUSTER_NAME=$1
SPEC_FILES_DIR=zzz-dev-scripts

delete_cluster() {
  go run -v ./cmd/kops -v10 delete cluster --name="$CLUSTER_NAME" --yes
  if [ $? != 0 ]; then
    echo "ERROR DELETING" "$1" "CLUSTER"
    exit 1
  fi
}

create_cluster() {
  go run -v ./cmd/kops -v10 create cluster --cloud=scaleway --zones=fr-par-1 --name="$CLUSTER_NAME" --networking=cilium --yes
  if [ $? != 0 ]; then
    echo "ERROR CREATING CLUSTER"
    exit 1
  fi
}

validate_cluster() {
  go run -v ./cmd/kops validate cluster --wait=10m
  if [ $? != 0 ]; then
    echo "COULD NOT VALIDATE CLUSTER WITHIN 10MIN"
    exit 1
  fi
}

update_cluster() {
  go run -v ./cmd/kops/ update cluster -v10 --name="$CLUSTER_NAME" --yes
  if [ $? != 0 ]; then
    echo "ERROR UPDATING CLUSTER"
    exit 1
  fi
}

replace_conf_file() {
  go run -v ./cmd/kops replace -f $SPEC_FILES_DIR/$CLUSTER_NAME-$1.yaml
  if [ $? != 0 ]; then
    echo "ERROR REPLACING CLUSTER SPEC FILE $SPEC_FILES_DIR/$CLUSTER_NAME-$1.yaml"
    exit 1
  fi
}

add_instance_group() {
  go run -v ./cmd/kops/ create instancegroup -v10 --name="$CLUSTER_NAME" "$1" --role="$2" --subnet=fr-par-1  --edit=false
    if [ $? != 0 ]; then
      echo "ERROR CREATING INSTANCE GROUP $1"
      exit 1
    fi
}

delete_instance_group() {
  go run -v ./cmd/kops/ delete instancegroup -v10 --name="$CLUSTER_NAME" "$1" --yes
  if [ $? != 0 ]; then
    echo "ERROR DELETING INSTANCE GROUP $1"
    exit 1
  fi
}

########################################################################################################################

if [ "$1" == "-d" ] || [ "$1" == "-c" ] || [ "$1" == "-am" ] || [ "$1" == "-rm" ] ; then
  echo "You forgot to give me a cluster name !"
  exit
fi

# DELETE PREVIOUS CLUSTER ?
if [ "$2" == "-d" ] ; then
  delete_cluster "previous"
  if [ "$3" == "" ]; then
    exit 0
  fi
fi

# CREATE CLUSTER ?
if [ "$2" == "-c" ] || [ "$3" == "-c" ] ; then
  create_cluster
fi

# ADD MASTERS ?
if [ "$2" == "-am" ] || [ "$3" == "-am" ] || [ "$4" == "-am" ] ; then
  validate_cluster
  replace_conf_file "extra_masters"
  add_instance_group "master2" "master"
  add_instance_group "master3" "master"
  update_cluster
  # REMOVE EXTRA MASTERS ?
  read -r -p "Are you ready to remove extra masters ? y or n" input
  if [[ $input == "y" ]] ; then
    replace_conf_file "simple"
    update_cluster
    delete_instance_group "master2"
    delete_instance_group "master3"
  fi
fi

# REMOVE EXTRA MASTERS ?
if [ "$2" == "-rm" ] || [ "$3" == "-rm" ] || [ "$4" == "-rm" ] || [ "$5" == "-rm" ] ; then
  if [[ $input == "y" ]] ; then
    replace_conf_file "simple"
    delete_instance_group "master2"
    delete_instance_group "master3"
    update_cluster
  fi
fi

printf '\a'

# DELETE CLUSTER ?
read -r -p "Are you ready to delete $CLUSTER_NAME ? y or n" input
if [[ $input == "y" ]]
then
  delete_cluster
fi