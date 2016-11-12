#!/bin/bash

# Copyright 2016 The Kubernetes Authors.
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
# with a custom version of the Nodeup binary compiled at runtime.
#
# Example usage
#
# S3_BUCKET="s3://my-dev-s3-state \
# CLUSTER_DOMAIN="mydomain.io" \
# CLUSTER_PREFIX="prefix" \
# NODEUP_BUCKET="s3-devel-bucket-name-store-nodeup" \
# ./dev-build.sh
#
###############################################################################

# This script assumes it's in $KOPS/hack

KOPS_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

[ -z "$S3_BUCKET" ] && echo "Need to set S3_BUCKET" && exit 1;
[ -z "$CLUSTER_DOMAIN" ] && echo "Need to set CLUSTER_DOMAIN" && exit 1;
[ -z "$CLUSTER_PREFIX" ] && echo "Need to set CLUSTER_PREFIX" && exit 1;
[ -z "$NODEUP_BUCKET" ] && echo "Need to set NODEUP_BUCKET" && exit 1;

# TODO make this better
CLUSTER_ENV="dev"
CLUSTER_REGION="us-west-2"
CLUSTER_SUBDOMAIN="aws-${CLUSTER_REGION}"
CLUSTER_NAME="${CLUSTER_PREFIX}.${CLUSTER_ENV}.${CLUSTER_SUBDOMAIN}.${CLUSTER_DOMAIN}"

# CLUSTER CONFIG
NODE_COUNT=${NODE_COUNT:-3}
NODE_ZONES=${NODE_ZONES:-"us-west-2a,us-west-2b,us-west-2c"}
NODE_SIZE=${NODE_SIZE:-m4.xlarge}
MASTER_ZONES=${MASTER_ZONES:-"us-west-2a,us-west-2b,us-west-2c"}
MASTER_SIZE=${MASTER_SIZE:-m4.large}

# NETWORK
TOPOLOGY=${TOPOLOGY:-private}
NETWORKING=${NETWORKING:-cni}

# NODEUP
NODEUP_OS="linux"
NODEUP_ARCH="amd64"

cd ../

GIT_VER=git-$(git describe --always)
# TODO check that GIT_VER worked
NODEUP_URL="https://s3-us-west-1.amazonaws.com/${NODEUP_BUCKET}/kops/${GIT_VER}/${NODEUP_OS}/${NODEUP_ARCH}/nodeup"

VERBOSITY=${VERBOSITY:-2}

make upload S3_BUCKET=s3://${NODEUP_BUCKET}

echo ==========
echo "Deleting cluster ${CLUSTER_NAME}"

kops delete cluster \
  --name $CLUSTER_NAME \
  --state $S3_BUCKET \
  -v $VERBOSITY \
  --yes

echo ==========
echo "Creating cluster ${CLUSTER_NAME}"

NODEUP_URL=${NODEUP_URL} kops create cluster \
  --name $CLUSTER_NAME \
  --state $S3_BUCKET \
  --node-count $NODE_COUNT \
  --zones $NODE_ZONES \
  --master-zones $MASTER_ZONES \
  --cloud aws \
  --dns-zone ${CLUSTER_SUBDOMAIN}.${CLUSTER_DOMAIN} \
  --node-size $NODE_SIZE \
  --master-size $MASTER_SIZE \
  --topology $TOPOLOGY \
  --networking $NETWORKING \
  -v $VERBOSITY \
  --yes

echo ==========
echo "Your k8s cluster ${CLUSTER_NAME}, awaits your bidding"
