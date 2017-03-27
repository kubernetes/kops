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

if [ $# -ne 1 ]; then
    echo Usage: vsphere_env [options]
    echo Options:
    echo -e "\t -s, --set   \t Set environment variables."
    echo -e "\t -u, --unset \t Unset environment variables."
    exit 1
fi

option="$1"
flag=0

case $option in
    -s | --set)
    # If set, coredns will be used for vsphere cloud provider.
    export VSPHERE_DNS=coredns

    # If set, this dns controller image will be used.
    # Leave this value unmodified if you are not building a new dns-controller image.
    export VSPHERE_DNSCONTROLLER_IMAGE=luomiao/dns-controller

    # S3 bucket that kops should use.
    export KOPS_STATE_STORE=s3://your-obj-store

    # AWS credentials
    export AWS_REGION=us-west-2
    export AWS_ACCESS_KEY_ID=something
    export AWS_SECRET_ACCESS_KEY=something

    # vSphere credentials
    export VSPHERE_USERNAME=administrator@vsphere.local
    export VSPHERE_PASSWORD=Admin!23

    # Set TARGET and TARGET_PATH to values where you want nodeup and protokube binaries to get copied.
    # This should be same location as set for NODEUP_URL and PROTOKUBE_IMAGE.
    export TARGET=jdoe@pa-dbc1131.eng.vmware.com
    export TARGET_PATH=/dbc/pa-dbc1131/jdoe/misc/kops/

    export NODEUP_URL=http://pa-dbc1131.eng.vmware.com/jdoe/misc/kops/nodeup/nodeup
    export PROTOKUBE_IMAGE=http://pa-dbc1131.eng.vmware.com/jdoe/misc/kops/protokube/protokube.tar.gz

    flag=1
    ;;
    -u | --unset)
    export VSPHERE_DNS=
    export VSPHERE_DNSCONTROLLER_IMAGE=
    export KOPS_STATE_STORE=
    export AWS_REGION=
    export AWS_ACCESS_KEY_ID=
    export AWS_SECRET_ACCESS_KEY=
    export VSPHERE_USERNAME=
    export VSPHERE_PASSWORD=
    export TARGET=
    export TARGET_PATH=
    export NODEUP_URL=
    export PROTOKUBE_IMAGE=

    flag=1
    ;;
    --default)
    echo Usage: vsphere_env [options]
    echo Options:
    echo -e "\t -s, --set   \t Set environment variables."
    echo -e "\t -u, --unset \t Unset environment variables."
    exit 1
    ;;
    *)
esac

if [[ $flag -ne 0 ]]; then
    echo "VSPHERE_DNS=${VSPHERE_DNS}"
    echo "VSPHERE_DNSCONTROLLER_IMAGE=${VSPHERE_DNSCONTROLLER_IMAGE}"
    echo "KOPS_STATE_STORE=${KOPS_STATE_STORE}"
    echo "AWS_REGION=${AWS_REGION}"
    echo "AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}"
    echo "AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}"
    echo "VSPHERE_USERNAME=${VSPHERE_USERNAME}"
    echo "VSPHERE_PASSWORD=${VSPHERE_PASSWORD}"
    echo "NODEUP_URL=${NODEUP_URL}"
    echo "PROTOKUBE_IMAGE=${PROTOKUBE_IMAGE}"
    echo "TARGET=${TARGET}"
    echo "TARGET_PATH=${TARGET_PATH}"
fi