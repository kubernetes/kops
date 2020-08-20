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

set -o errexit
set -o nounset
set -o pipefail

. "$(dirname "${BASH_SOURCE[0]}")/common.sh"

cd "${KOPS_ROOT}"

# Accept an optional argument overriding the package to update
PKG="${1:-./...}"

# Update gobindata to reflect any yaml changes
make kops-gobindata

# Don't override variables that are commonly used in dev, but shouldn't be in our tests
unset KOPS_BASE_URL DNSCONTROLLER_IMAGE KOPSCONTROLLER_IMAGE KUBE_APISERVER_HEALTHCHECK_IMAGE KOPS_FEATURE_FLAGS
unset AWS_ACCESS_KEY_ID AWS_REGION AWS_SECRET_ACCESS_KEY AWS_SESSION_TOKEN CNI_VERSION_URL DNS_IGNORE_NS_CHECK DO_ACCESS_TOKEN GOOGLE_APPLICATION_CREDENTIALS
unset KOPS_CLUSTER_NAME KOPS_RUN_OBSOLETE_VERSION KOPS_STATE_STORE KOPS_STATE_S3_ACL KUBE_API_VERSIONS NODEUP_URL OPENSTACK_CREDENTIAL_FILE PROTOKUBE_IMAGE SKIP_PACKAGE_UPDATE
unset SKIP_REGION_CHECK S3_ACCESS_KEY_ID S3_ENDPOINT S3_REGION S3_SECRET_ACCESS_KEY


# Run the tests in "autofix mode"
HACK_UPDATE_EXPECTED_IN_PLACE=1 go test "${PKG}" -count=1
