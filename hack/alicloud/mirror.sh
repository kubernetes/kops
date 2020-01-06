#!/usr/bin/env bash

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

IMAGES_FILE='hack/alicloud/required-images.txt'
repos=$(grep -v ^# $IMAGES_FILE | cut -d: -f1 | sort -u)

ACR_DN='registry.cn-shanghai.aliyuncs.com/bkcn'

function need_trim() {
  s=$1
  for i in $NO_TRIM
  do
    if [ "${s/$i/}" == "$s" ]; then
      continue
    else
      return 1
    fi
  done
  return 0
}

function pull_and_push(){
  origimg="$1"
  if need_trim $origimg; then
    echo "$origimg needs triming"
    # strip off the prefix
    img=${origimg/gcr.io\/google_containers\//}
    img=${img/k8s.gcr.io\//}
    target_img="$ACR_DN/${img//\//-}"
  else
    echo "$origimg does not need triming"
    target_img="$ACR_DN/$origimg"
  fi

  docker pull $origimg
  echo "tagging $origimg to $target_img"
  docker tag $origimg $target_img
  echo "[PUSH] remote image not exists or digests not match, pushing $target_img"
  docker push $target_img
}

for r in ${repos[@]}
do
  if need_trim $r; then
    # strip off the prefix
    r=${r/gcr.io\/google_containers\//}
    r=${r/k8s.gcr.io\//}
    r=${r//\//-}
  fi
done

images=$(grep -v ^# $IMAGES_FILE)
for i in ${images[@]}
do
  pull_and_push $i
done
