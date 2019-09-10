#!/bin/bash -ex

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


mkdir -p /go
export GOPATH=/go

mkdir -p /go/src/k8s.io
ln -s /src/ /go/src/k8s.io/kops

cd /go/src/k8s.io/kops/
make dns-controller-gocode

mkdir -p /src/.build/artifacts/
cp /go/bin/dns-controller /src/.build/artifacts/

chown -R $HOST_UID:$HOST_GID /src/.build/artifacts
