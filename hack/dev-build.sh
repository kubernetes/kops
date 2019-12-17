#!/usr/bin/env bash

# Copyright 2017 The Kubernetes Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.


###############################################################################
#
# dev-build.sh
#
# Convenience script for developing kops AND nodeup.
#
# This script (by design) will handle building a full kops cluster in AWS,
# with a custom version of the nodeup, protokube and dnscontroller.
#
# This script and Makefile uses aws client
# https://aws.amazon.com/cli/
# and make sure you `aws configure`
#
# # Example usage
#
# KOPS_STATE_STORE="s3://my-dev-s3-state" \
# CLUSTER_NAME="fullcluster.name.mydomain.io" \
# NODEUP_BUCKET="s3-devel-bucket-name-store-nodeup" \
# IMAGE="kope.io/k8s-1.6-debian-jessie-amd64-hvm-ebs-2017-05-02" \
# ./dev-build.sh
#
# # TLDR;
# 1. setup dns in route53
# 2. create s3 buckets - state store and nodeup bucket
# 3. set zones appropriately, you need 3 zones in a region for HA
# 4. run script
# 5. find bastion to ssh into (look in ELBs)
# 6. use ssh-agent and ssh -A
# 7. your pem will be the access token
# 8. user is admin, and the default is debian
#
# # For more details see:
#
# https://github.com/kubernetes/kops/blob/master/docs/getting_started/aws.md
#
###############################################################################

KOPS_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

#
# Check that required binaries are installed
#
command -v make >/dev/null 2>&1 || { echo >&2 "I require make but it's not installed.  Aborting."; exit 1; }
command -v go >/dev/null 2>&1 || { echo >&2 "I require go but it's not installed.  Aborting."; exit 1; }
command -v docker >/dev/null 2>&1 || { echo >&2 "I require docker but it's not installed.  Aborting."; exit 1; }
command -v aws >/dev/null 2>&1 || { echo >&2 "I require aws cli but it's not installed.  Aborting."; exit 1; }

#
# Check that expected vars are set
#
[ -z "$KOPS_STATE_STORE" ] && echo "Need to set KOPS_STATE_STORE" && exit 1;
[ -z "$CLUSTER_NAME" ] && echo "Need to set CLUSTER_NAME" && exit 1;
[ -z "$NODEUP_BUCKET" ] && echo "Need to set NODEUP_BUCKET" && exit 1;
[ -z "$IMAGE" ] && echo "Need to set IMAGE or use the image listed here https://github.com/kubernetes/kops/blob/master/channels/stable" && exit 1;

# Cluster config
NODE_COUNT=${NODE_COUNT:-3}
NODE_ZONES=${NODE_ZONES:-"us-west-2a,us-west-2b,us-west-2c"}
NODE_SIZE=${NODE_SIZE:-m4.xlarge}
MASTER_ZONES=${MASTER_ZONES:-"us-west-2a,us-west-2b,us-west-2c"}
MASTER_SIZE=${MASTER_SIZE:-m4.large}
KOPS_CREATE=${KOPS_CREATE:-yes}

# NETWORK
TOPOLOGY=${TOPOLOGY:-private}
NETWORKING=${NETWORKING:-weave}

# How verbose go logging is
VERBOSITY=${VERBOSITY:-10}

cd $KOPS_DIRECTORY/..

GIT_VER=git-$(git describe --always)
[ -z "$GIT_VER" ] && echo "we do not have GIT_VER something is very wrong" && exit 1;

echo ==========
echo "Starting build"

# removing CI=1 because it forces a new upload every time
# export CI=1
make && UPLOAD_DEST=s3://${NODEUP_BUCKET} make upload

# removing make test since it relies on the files in the bucket
# && make test

KOPS_VERSION=$(kops version --short)
KOPS_BASE_URL="http://${NODEUP_BUCKET}.s3.amazonaws.com/kops/${KOPS_VERSION}/"

echo "KOPS_BASE_URL=${KOPS_BASE_URL}"
echo "NODEUP_URL=${KOPS_BASE_URL}linux/amd64/nodeup"

echo ==========
echo "Deleting cluster ${CLUSTER_NAME}. Elle est finie."

kops delete cluster \
  --name $CLUSTER_NAME \
  --state $KOPS_STATE_STORE \
  -v $VERBOSITY \
  --yes

echo ==========
echo "Creating cluster ${CLUSTER_NAME}"

kops_command="NODEUP_URL=${KOPS_BASE_URL}linux/amd64/nodeup KOPS_BASE_URL=${KOPS_BASE_URL} kops create cluster --name $CLUSTER_NAME --state $KOPS_STATE_STORE --node-count $NODE_COUNT --zones $NODE_ZONES --master-zones $MASTER_ZONES --node-size $NODE_SIZE --master-size $MASTER_SIZE -v $VERBOSITY --image $IMAGE --channel alpha --topology $TOPOLOGY --networking $NETWORKING"

if [[ $TOPOLOGY == "private" ]]; then
  kops_command+=" --bastion='true'"
fi

if [ -n "${KOPS_FEATURE_FLAGS+x}" ]; then
  kops_command=KOPS_FEATURE_FLAGS="${KOPS_FEATURE_FLAGS}" $kops_command
fi

if [[ $KOPS_CREATE == "yes" ]]; then
  kops_command="$kops_command --yes"
fi

eval $kops_command

echo ==========
echo "Your k8s cluster ${CLUSTER_NAME}, awaits your bidding."
