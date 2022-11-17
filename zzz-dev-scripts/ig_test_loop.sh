#!/usr/bin/bash

CLUSTER_NAME=$1

if [ "$2" == "-d" ]; then
  go run -v ./cmd/kops -v10 delete cluster --name="$CLUSTER_NAME" --yes
  if [ $? != 0 ]; then
    echo "ERROR DELETING PREVIOUS CLUSTER"
    exit
  fi
fi
go run -v ./cmd/kops -v10 create cluster --cloud=scaleway --zones=fr-par-1 --name="$CLUSTER_NAME" --networking=calico --yes
if [ $? != 0 ]; then
  echo "ERROR CREATING CLUSTER"
  exit
fi
go run -v ./cmd/kops go run -v ./cmd/kops edit ig control-plane-fr-par-1
if [ $? != 0 ]; then
  echo "ERROR GROWING INSTANCE GROUP"
  exit
fi
go run -v ./cmd/kops/ update cluster -v10 --name="$CLUSTER_NAME" --yes
if [ $? != 0 ]; then
  echo "ERROR UPDATING CLUSTER"
  exit
fi
go run -v ./cmd/kops go run -v ./cmd/kops edit ig control-plane-fr-par-1
if [ $? != 0 ]; then
  echo "ERROR SHRINKING INSTANCE GROUP"
  exit
fi
go run -v ./cmd/kops/ update cluster -v10 --name="$CLUSTER_NAME" --yes
if [ $? != 0 ]; then
  echo "ERROR UPDATING CLUSTER"
  exit
fi
go run -v ./cmd/kops/ create instancegroup -v10 --name="$CLUSTER_NAME" master2 --role=master --subnet=fr-par-1 --edit=false
if [ $? != 0 ]; then
  echo "ERROR CREATING INSTANCE GROUP MASTER 2"
  exit
fi
go run -v ./cmd/kops/ create instancegroup -v10 --name="$CLUSTER_NAME" master3 --role=master --subnet=fr-par-1 --edit=false
if [ $? != 0 ]; then
  echo "ERROR CREATING INSTANCE GROUP MASTER 3"
  exit
fi
go run -v ./cmd/kops/ update cluster -v10 --name="$CLUSTER_NAME" --yes
if [ $? != 0 ]; then
  echo "ERROR UPDATING CLUSTER"
  exit
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