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


HEADER=$(cat hack/boilerplate/boilerplate.sh.txt | sed 's/YEAR/2016/')

SCRIPTS=$(hack/verify-boilerplate.sh | awk  '{ print $6 }' | grep sh$)
for i in ${SCRIPTS}
do
	:
	value=$(<$i)
	if [[ $value == *"# Copyright"* ]]
	then
		  echo "Bad header in $i"
		  continue
	fi
	echo -e "${HEADER}\n\n${value}" > $i
done

HEADER=$(cat hack/boilerplate/boilerplate.go.txt | sed 's/YEAR/2016/')

SCRIPTS=$(hack/verify-boilerplate.sh | awk  '{ print $6 }' | grep go$)
for i in ${SCRIPTS}
do
	:
	value=$(<$i)
	if [[ $value == *"# Copyright"* ]]
	then
		  echo "Bad header in $i"
		  continue
	fi
	echo -e "${HEADER}\n\n${value}" > $i
done


