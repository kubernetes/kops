#!/bin/sh -ex

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

rm -f /dist/runc*

cp /go/src/github.com/docker/runc/runc /dist/runc-17.03.2

(sha1sum /dist/runc-17.03.2 | cut -d' ' -f1) > /dist/runc-17.03.2.sha1
