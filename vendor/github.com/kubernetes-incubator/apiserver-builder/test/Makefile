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

.PHONY: all test build generate docs cmds clean cleangenerated cleandocs skeleton

all: test

test: build
	go test ./pkg/...
	bash -c "find pkg/apis/ -name apiserver.local.config | xargs rm -rf"

skeleton: cmds
	apiserver-boot init repo --domain sample.kubernetes.io
	apiserver-boot create group version resource --group storage --version v1 --kind VolumeClaim
	apiserver-boot create group version resource --group storage --version v1 --kind SnapshotClaim
	apiserver-boot create group version resource --group storage --version v1 --kind Volume --non-namespaced
	apiserver-boot create group version resource --group storage --version v1 --kind Snapshot --non-namespaced

build: cmds skeleton
	apiserver-boot build executables

# Build docs
docs: cleandocs build
	apiserver-boot build docs --server bin/apiserver

# Build the tools
# Note: These are available in the apiserver-builder release tar.gz
cmds:
	./build-tools.sh

# Clean up functions
clean: cleangenerated cleandocs

cleangenerated: cmds
	apiserver-boot build generated clean

cleandocs: cmds
	apiserver-boot build docs clean
