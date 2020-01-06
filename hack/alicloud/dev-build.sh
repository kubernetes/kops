#!/usr/bin/env bash

# Copyright 2019 The Kubernetes Authors.
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
# dev-build-alicloud.sh
#
# Convenience script for developing kops AND nodeup on Alicloud.
#
# This script (by design) will handle building a full kops cluster in Alicloud,
# with a custom version of the nodeup, protokube and dnscontroller.
#
# This script and Makefile uses aliyun client
# https://github.com/aliyun/aliyun-cli
# and make sure you `aliyun configure`
#
# # Example usage
#
# KOPS_STATE_STORE="oss://my-dev-oss-state" \
# CLUSTER_NAME="fullcluster.name.k8s.local" \
# NODEUP_BUCKET="oss-devel-bucket-name-store-nodeup" \
# IMAGE="m-xxxxxxxxxxxxxxxxxxxxxxx" \
# ./hack/alicloud/dev-build.sh
#
# # TLDR;
# 1. create oss buckets - state store and nodeup bucket
# 2. set zones appropriately, you need 3 zones in a region for HA
# 3. run script
# 4. use ssh-agent and ssh -A
# 5. your pem will be the access token
#
# # For more details see:
#
# https://github.com/kubernetes/kops/blob/master/docs/development/alicloud.md
#
###############################################################################

KOPS_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

#
# Check that required binaries are installed
#
command -v make >/dev/null 2>&1 || { echo >&2 "I require make but it's not installed.  Aborting."; exit 1; }
command -v go >/dev/null 2>&1 || { echo >&2 "I require go but it's not installed.  Aborting."; exit 1; }
command -v docker >/dev/null 2>&1 || { echo >&2 "I require docker but it's not installed.  Aborting."; exit 1; }
command -v aliyun >/dev/null 2>&1 || { echo >&2 "I require aliyun cli but it's not installed.  Aborting."; exit 1; }

#
# Check that expected vars are set
#
[ -z "$KOPS_STATE_STORE" ] && echo "Need to set KOPS_STATE_STORE" && exit 1;
[ -z "$CLUSTER_NAME" ] && echo "Need to set CLUSTER_NAME" && exit 1;
[ -z "$NODEUP_BUCKET" ] && echo "Need to set NODEUP_BUCKET" && exit 1;
[ -z "$IMAGE" ] && echo "Need to set IMAGE or use the image listed here https://github.com/kubernetes/kops/blob/master/channels/stable" && exit 1;

# Cluster config
KUBERNETES_VERSION=1.14.10
NODE_COUNT=${NODE_COUNT:-3}
NODE_ZONES=${NODE_ZONES:-"cn-shanghai-e,cn-shanghai-f,cn-shanghai-g"}
NODE_SIZE=${NODE_SIZE:-ecs.g6.large}
MASTER_ZONES=${MASTER_ZONES:-"cn-shanghai-e,cn-shanghai-f,cn-shanghai-g"}
MASTER_SIZE=${MASTER_SIZE:-ecs.g6.large}
KOPS_CREATE=${KOPS_CREATE:-yes}

# NETWORK
TOPOLOGY=${TOPOLOGY:-private}
NETWORKING=${NETWORKING:-flannel}

# How verbose go logging is
VERBOSITY=${VERBOSITY:-10}

cd $KOPS_DIRECTORY/..

GIT_VER=git-$(git describe --always)
[ -z "$GIT_VER" ] && echo "we do not have GIT_VER something is very wrong" && exit 1;

echo ==========
echo "Starting build"

# removing CI=1 because it forces a new upload every time
# export CI=1
make && OSS_BUCKET=oss://${NODEUP_BUCKET} make oss-upload
if [[ $? -ne 0 ]]; then
  exit 1
fi

# removing make test since it relies on the files in the bucket
# && make test

KOPS_VERSION=$(kops version --short)
KOPS_BASE_URL="https://${NODEUP_BUCKET}.${OSS_REGION}.aliyuncs.com/kops/${KOPS_VERSION}/"

echo "KOPS_BASE_URL=${KOPS_BASE_URL}"
echo "NODEUP_URL=${KOPS_BASE_URL}linux/amd64/nodeup"

echo ==========
echo "Deleting cluster ${CLUSTER_NAME}..."

kops delete cluster \
  --name $CLUSTER_NAME \
  --state $KOPS_STATE_STORE \
  -v $VERBOSITY \
  --yes

echo ==========
echo "Creating cluster ${CLUSTER_NAME}..."

kops_command="export NODEUP_URL=${KOPS_BASE_URL}linux/amd64/nodeup; export KOPS_BASE_URL=${KOPS_BASE_URL}; kops create cluster --cloud=alicloud --name $CLUSTER_NAME --state $KOPS_STATE_STORE --node-count $NODE_COUNT --zones $NODE_ZONES --master-zones $MASTER_ZONES --node-size $NODE_SIZE --master-size $MASTER_SIZE -v $VERBOSITY --image $IMAGE --channel alpha --topology $TOPOLOGY --networking $NETWORKING --kubernetes-version $KUBERNETES_VERSION"

# bastion is not supported in Alicloud yet
# if [[ $TOPOLOGY == "private" ]]; then
#   kops_command+=" --bastion='true'"
# fi

if [ -n "${KOPS_FEATURE_FLAGS+x}" ]; then
  kops_command="export KOPS_FEATURE_FLAGS=${KOPS_FEATURE_FLAGS}; $kops_command"
  echo $kops_command
fi

if [[ $KOPS_CREATE == "yes" ]]; then
  kops_command="$kops_command --yes"
fi

eval $kops_command

echo ==========
echo "Your k8s cluster ${CLUSTER_NAME}, awaits your bidding."
