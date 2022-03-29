#!/usr/bin/env bash

# Copyright 2022 The Kubernetes Authors.
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

if ! command -v gsutil &> /dev/null; then
    curl https://dl.google.com/dl/cloudsdk/channels/rapid/google-cloud-sdk.tar.gz -o /tmp/google-cloud-sdk.tar.gz
    tar xzf /tmp/google-cloud-sdk.tar.gz -C /
    rm /tmp/google-cloud-sdk.tar.gz
    /google-cloud-sdk/install.sh \
        --bash-completion=false \
        --usage-reporting=false \
        --quiet
    ln -s /google-cloud-sdk/bin/gcloud /usr/local/bin/gcloud
    ln -s /google-cloud-sdk/bin/gsutil /usr/local/bin/gsutil
    gcloud info
    gcloud config list
    gcloud auth list
fi
