#!/usr/bin/env bash

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


###############################################################################
#
# new-iam-user.sh
#
# Convenience script adding a new IAM user to an existing AWS account.
#
# WARNING: This script will return secrets!
#
###############################################################################

usage(){
    echo "sh new-iam-user.sh <group-name> <user-name>"
    exit 0
}

if [ -z "$1" ]; then
    usage
fi

if [ -z "$2" ]; then
    usage
fi

command -v aws >/dev/null 2>&1 || { echo >&2 "The aws cli is required to run this script."; exit 1; }

GROUP=$1
USER=$2

aws iam create-group --group-name ${GROUP}

export arns="
arn:aws:iam::aws:policy/AmazonEC2FullAccess
arn:aws:iam::aws:policy/AmazonRoute53FullAccess
arn:aws:iam::aws:policy/AmazonS3FullAccess
arn:aws:iam::aws:policy/IAMFullAccess
arn:aws:iam::aws:policy/AmazonVPCFullAccess"

for arn in $arns; do aws iam attach-group-policy --policy-arn "$arn" --group-name ${GROUP}; done

aws iam create-user --user-name ${USER}

aws iam add-user-to-group --user-name ${USER} --group-name ${GROUP}

aws iam create-access-key --user-name ${USER}