#!/bin/bash

# Copyright The Kubernetes Authors.
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

in=$1
out=$2

# Preset 6 is the knee of the curve for the nodeup binary: -9e is only ~22KB smaller but needs
# 65MiB rather than 9MiB to decompress on the node. Stay single-threaded so that both branches
# below produce identical bytes; threaded xz splits the input into blocks and compresses worse.
if ( command -v xz > /dev/null ); then
  xz -6 -T1 -c "$in" > "$out"
elif ( command -v python3 > /dev/null ); then
  # The release build image has no xz binary but does have python3, whose lzma module is
  # liblzma, so this produces the same stream as xz -6.
  python3 -c 'import lzma,sys; sys.stdout.buffer.write(lzma.compress(open(sys.argv[1],"rb").read(), preset=6))' "$in" > "$out"
else
  echo "Neither xz nor python3 command is available"
  exit 1
fi
