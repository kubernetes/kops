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

set -e
set -x

mkdir -p /go/src/github.com/docker
cd /go/src/github.com/docker

git clone https://github.com/docker/runc
cd runc

git checkout 54296cf40ad8143b62dbcaa1d90e520a2136ddfe

# Apply CVE-2019-5736 patch (backported)
cat /CVE-2019-5736.patch | git apply -v --index

git config user.email "kops@kubernetes.io"
git config user.name "kops"

git commit -m "Applied CVE patch"

GOPATH=/go make BUILDTAGS="seccomp apparmor selinux" static

ls
