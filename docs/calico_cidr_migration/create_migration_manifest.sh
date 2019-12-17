#!/bin/bash

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
# create_migration_manifest.sh
#
# Script that returns a templated Calico CIDR migration manifest file.
#
###############################################################################

set -e

command -v jq >/dev/null 2>&1 || { echo >&2 "jq is required to run this script."; exit 1; }
command -v kops >/dev/null 2>&1 || { echo >&2 "kops is required to run this script."; exit 1; }

[ -z "$NAME" ] && echo "Please set NAME to the name of your cluster you wish to perform this migration against." && exit 1;

export MIGRATION_TEMPLATE="jobs.yaml.template"
export MIGRATION_MANIFEST="jobs.yaml"
export NON_MASQUERADE_CIDR="`kops get cluster $NAME -o json --full | jq .spec.nonMasqueradeCIDR --raw-output`"
export POD_CIDR="`kops get cluster $NAME -o json --full | jq .spec.kubeControllerManager.clusterCIDR --raw-output`"
export IS_CROSS_SUBNET="`kops get cluster $NAME -o json --full | jq .spec.networking.calico.crossSubnet --raw-output`"

cp ${MIGRATION_TEMPLATE} ${MIGRATION_MANIFEST}

if [ "$IS_CROSS_SUBNET" = "true" ]; then
    echo "ipip mode is set to 'cross-subnet'. Honouring in migration manifest."
else
    echo "ipip mode is set to 'Always'. Honouring in migration manifest."
    sed -i "/mode: cross-subnet/d" ${MIGRATION_MANIFEST}
fi

sed -i -e "s@{{NON_MASQUERADE_CIDR}}@${NON_MASQUERADE_CIDR}@g" ${MIGRATION_MANIFEST}
sed -i -e "s@{{POD_CIDR}}@${POD_CIDR}@g" ${MIGRATION_MANIFEST}

echo "jobs.yaml created. Please run: "
echo "kubectl apply -f jobs.yaml"
