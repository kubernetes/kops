#!/usr/bin/env bash

# Copyright 2023 The Kubernetes Authors.
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

set -e
set -x

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "${REPO_ROOT}"

if [[ -z "${IMAGE_PREFIX}" ]]; then
  IMAGE_PREFIX="${USER}/"
fi

IMAGE_TAG=$(date +%Y%m%d%H%M%S)

# Build the controller image
KO_DOCKER_REPO="${IMAGE_PREFIX}kops-controller" go run github.com/google/ko@v0.14.1 \
  build --tags "${IMAGE_TAG}" --platform=linux/amd64,linux/arm64 --bare ./cmd/kops-controller/

# Update the image and bounce the pods
kubectl set image -n kube-system daemonset/kops-controller "*=${IMAGE_PREFIX}kops-controller:${IMAGE_TAG}"
kubectl delete pod -n kube-system -l k8s-addon=kops-controller.addons.k8s.io