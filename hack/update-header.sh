#!/bin/bash

# Copyright 2016 The Kubernetes Authors.
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

KUBE_ROOT=$(dirname "${BASH_SOURCE}")/..
BAD_HEADERS=$(${KUBE_ROOT}/hack/verify-boilerplate.sh | awk '{ print $6}')
FORMATS="sh go Makefile Dockerfile"

for i in ${FORMATS}
do
	:
	for j in ${BAD_HEADERS}
	do
		:
	        HEADER=$(cat hack/boilerplate/boilerplate.${i}.txt | sed 's/YEAR/2016/')
			value=$(<${j})
			if [[ "$j" != *$i ]]
            then
                continue
            fi

			if [[ ${value} == *"# Copyright"* ]]
			then
				echo "Bad header in ${j} ${i}"
			else
				text="$HEADER

$value"
				echo ${j}
				echo "$text" > ${j}
			fi
	done
done
