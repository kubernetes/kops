#!/bin/bash
###############################################################################
#
# dev-build.sh
#
# Convenience script for developing kops AND nodeup.
#
# This script (by design) will handle building a full kops cluster in AWS,
# with a custom version of the Nodeup binary compiled at runtime.
#
#
###############################################################################

# This script assumes it's in $KOPS/hack
cd ../

KOPS_DIRECTORY="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
GIT_VER=git-$(git describe --always)
VERBOSITY=10

# STATE_STORE
export S3_BUCKET="s3://oscar-ai-k8s"

# CLUSTER NAME
CLUSTER_PREFIX="k8s-001"
CLUSTER_ENV="dev"
CLUSTER_REGION="us-west-2"
CLUSTER_SUBDOMAIN="aws-${CLUSTER_REGION}"
CLUSTER_DOMAIN="mydomain.com"
CLUSTER_NAME="${CLUSTER_PREFIX}.${CLUSTER_ENV}.${CLUSTER_SUBDOMAIN}.${CLUSTER_DOMAIN}"

# CLUSTER CONFIG
NODE_COUNT=3
NODE_ZONES="us-west-2a,us-west-2b,us-west-2c"
NODE_SIZE="m4.large"
MASTER_ZONES="us-west-2a,us-west-2b,us-west-2c"
MASTER_SIZE="m4.xlarge"

# NETWORK
TOPOLOGY="private"
NETWORKING="cni"

# NODEUP
NODEUP_OS="linux"
NODEUP_ARCH="amd64"
NODEUP_BUCKET="kops-devel"
export NODEUP_URL="https://${NODEUP_BUCKET}.s3-us-west-1.com/${NODEUP_BUCKET}/kops/${GIT_VER}/${NODEUP_OS}/${NODEUP_ARCH}/nodeup"
make version-dist
make
aws s3 sync --acl public-read .build/upload/ s3://${NODEUP_BUCKET}



kops delete cluster \
  --name $CLUSTER_NAME \
  --state $S3_BUCKET \
  -v $VERBOSITY \
  --yes

kops create cluster \
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