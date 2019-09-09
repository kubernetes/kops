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

rm -f /utils.tar.gz
rm -rf /utils

mkdir -p /utils
cp /socat/socat-*/debian/socat/usr/bin/socat /utils/socat
cp /conntrack/conntrack-*/debian/conntrack/usr/sbin/conntrack /utils/conntrack
#(sha1sum /utils/socat | cut -d' ' -f1) > /utils/socat.sha1

tar cvfz /utils.tar.gz /utils

cp /utils.tar.gz /dist/utils.tar.gz
(sha1sum /dist/utils.tar.gz | cut -d' ' -f1) > /dist/utils.tar.gz.sha1
(sha256sum /dist/utils.tar.gz | cut -d' ' -f1) > /dist/utils.tar.gz.sha256

