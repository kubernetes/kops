#!/usr/bin/bash

CLUSTER_NAME=$1
SPEC_FILES_DIR=zzz-dev-scripts

if [ "$1" == "-d" ] || [ "$1" == "-c" ] || [ "$1" == "-am" ] || [ "$1" == "-rm" ] ; then
  echo "You forgot to give me a cluster name !"
  exit
fi

# DELETE PREVIOUS CLUSTER ?
if [ "$2" == "-d" ] || [ "$3" == "-d" ] || [ "$4" == "-d" ] ; then
  go run -v ./cmd/kops -v10 delete cluster --name="$CLUSTER_NAME" --yes
  if [ $? != 0 ]; then
    echo "ERROR DELETING PREVIOUS CLUSTER"
    exit
  fi
fi

# CREATE CLUSTER ?
if [ "$2" == "-c" ] || [ "$3" == "-c" ] || [ "$4" == "-c" ] ; then
  go run -v ./cmd/kops -v10 create cluster --cloud=scaleway --zones=fr-par-1 --name="$CLUSTER_NAME" --networking=cilium --yes
  if [ $? != 0 ]; then
    echo "ERROR CREATING CLUSTER"
    exit
  fi
fi

# ADD MASTERS ?
if [ "$2" == "-am" ] || [ "$3" == "-am" ] || [ "$4" == "-am" ] ; then
  go run -v ./cmd/kops replace -f "$SPEC_FILES_DIR/$CLUSTER_NAME"_extra_masters.yaml
  if [ $? != 0 ]; then
    echo "ERROR REPLACING CLUSTER SPEC FILE"
    exit
  fi
  go run -v ./cmd/kops/ create instancegroup -v10 --name="$CLUSTER_NAME" master2 --role=master --subnet=fr-par-1 --edit=false
  if [ $? != 0 ]; then
    echo "ERROR CREATING INSTANCE GROUP MASTER 2"
    exit
  fi
  go run -v ./cmd/kops/ create instancegroup -v10 --name="$CLUSTER_NAME" master3 --role=master --subnet=fr-par-1 --edit=false
  if [ $? != 0 ]; then
    echo "ERROR CREATING INSTANCE GROUP MASTER 3"Nkfouuf!ToFo
        exit
  fi
  go run -v ./cmd/kops/ update cluster -v10 --name="$CLUSTER_NAME" --yes
  if [ $? != 0 ]; then
    echo "ERROR UPDATING CLUSTER"
    exit
  fi

  # REMOVE EXTRA MASTERS ?
  read -r -p "Are you ready to remove extra masters ? y or n" input
  if [[ $input == "y" ]] ; then
    go run -v ./cmd/kops/ delete instancegroup -v10 --name=$CLUSTER_NAME master2 --yes
    go run -v ./cmd/kops/ delete instancegroup -v10 --name=$CLUSTER_NAME master3 --yes
    go run -v ./cmd/kops replace -f "$SPEC_FILES_DIR/$CLUSTER_NAME"_simple.yaml
    go run -v ./cmd/kops/ update cluster -v10 --name=$CLUSTER_NAME --yes
  fi
fi

printf '\a'

read -r -p "Are you ready to delete $CLUSTER_NAME ? y or n" input
if [[ $input == "y" ]]
then
  go run -v ./cmd/kops -v10 delete cluster --name="$CLUSTER_NAME" --yes
  if [ $? != 0 ]; then
    echo "ERROR DELETING CLUSTER"
    exit
  fi
fi