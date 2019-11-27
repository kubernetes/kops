#!/bin/bash

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

set -o errexit
set -o nounset
set -o pipefail

P_OPERATOR_VERSION=${1:-"v0.26.0"}
P_OPERATOR_ADDON_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd ${P_OPERATOR_ADDON_DIR}
git clone -b ${P_OPERATOR_VERSION} --depth 1 https://github.com/coreos/prometheus-operator
cp prometheus-operator/bundle.yaml ${P_OPERATOR_VERSION}.yaml
mkdir tmp
cp prometheus-operator/contrib/kube-prometheus/manifests/* tmp
for i in `ls tmp`
do
  echo "---" >> ${P_OPERATOR_VERSION}.yaml
  cat tmp/$i >> ${P_OPERATOR_VERSION}.yaml
done

rm -rf ${P_OPERATOR_ADDON_DIR}/prometheus-operator ${P_OPERATOR_ADDON_DIR}/tmp/
cd -
