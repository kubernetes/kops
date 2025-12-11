#!/bin/bash

# Copyright 2024 The Kubernetes Authors.
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

set -x

if [[ -z "${VERSION:-}" ]]; then
    echo "Must specify VERSION"
    exit 1
fi

GITHUB_USER=$(gh api user | jq -r '.login')

kpromo="go run sigs.k8s.io/promo-tools/v4/cmd/kpromo@v4.0.4"

WORKDIR=$(mktemp -d -t workflow.XXXXXXXX)
echo "WORKDIR is ${WORKDIR}"

git clone https://github.com/kubernetes/k8s.io ${WORKDIR}/k8s.io --depth=8
cd ${WORKDIR}/k8s.io
gh repo set-default kubernetes/k8s.io

git remote add fork https://github.com/${GITHUB_USER}/k8s.io

# Promote images
cd ${WORKDIR}/k8s.io

git checkout main
git pull origin main
git checkout -b kops_images_${VERSION}

echo "" >> registry.k8s.io/images/k8s-staging-kops/images.yaml
echo "# ${VERSION}" >> registry.k8s.io/images/k8s-staging-kops/images.yaml
${kpromo} cip run --snapshot gcr.io/k8s-staging-kops --snapshot-tag ${VERSION} >> registry.k8s.io/images/k8s-staging-kops/images.yaml

git add registry.k8s.io/images/k8s-staging-kops/images.yaml
git commit -m "Promote kOps $VERSION images"

git push fork --force
gh pr create --fill --base main --head ${GITHUB_USER}:kops_images_${VERSION}


# Promote binary artifacts
cd ${WORKDIR}/k8s.io

git checkout main
git pull origin main
git checkout -b kops_artifacts_${VERSION}

rm -rf ./k8s-staging-kops/kops/releases
mkdir -p ./k8s-staging-kops/kops/releases/${VERSION}/
gsutil rsync -r  gs://k8s-staging-kops/kops/releases/${VERSION}/ ./k8s-staging-kops/kops/releases/${VERSION}/

${kpromo} manifest files --src k8s-staging-kops/kops/releases/ >> artifacts/manifests/k8s-staging-kops/${VERSION}.yaml

git add artifacts/manifests/k8s-staging-kops/${VERSION}.yaml
git commit -m "Promote kOps $VERSION binary artifacts"

git push fork --force
gh pr create --fill --base main --head ${GITHUB_USER}:kops_artifacts_${VERSION}
