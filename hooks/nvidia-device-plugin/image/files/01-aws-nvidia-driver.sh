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

#!/bin/bash
set -euo pipefail
set -x

#################################################
# Settings

# A place on the host machine to cache 1.6GB+ downloads in-between reboots.

CACHE_DIR=/nvidia-device-plugin

# AWS Instance Types to Nvidia Card Mapping (cut and pasted from AWS docs)
# Load the correct driver for the correct instance type
#   Instances  Product Type  Product Series  Product
#   G2         GRID          GRID Series     GRID K520 (deprecated)
#   G3         Tesla         M-Series        M-60
#   G3S        Tesla         M-Series        M-60
#   G4         Tesla         T-Series        T-4
#   P2         Tesla         K-Series        K-80
#   P3         Tesla         V-Series        V100
# http://www.nvidia.com/Download/index.aspx
declare -A class_to_driver_file
declare -A class_to_driver_checksum
case $CUDA_VERSION in
    9.1)
        class_to_driver_file=( \
            ["g3"]="http://us.download.nvidia.com/tesla/390.46/NVIDIA-Linux-x86_64-390.46.run" \
            ["g3s"]="http://us.download.nvidia.com/tesla/390.46/NVIDIA-Linux-x86_64-390.46.run" \
            ["g4dn"]="http://us.download.nvidia.com/tesla/418.87/NVIDIA-Linux-x86_64-390.46.00.run" \
            ["p2"]="http://us.download.nvidia.com/tesla/390.46/NVIDIA-Linux-x86_64-390.46.run" \
            ["p3"]="http://us.download.nvidia.com/tesla/390.46/NVIDIA-Linux-x86_64-390.46.run" \
        )

        class_to_driver_checksum=( \
            ["g3"]="57569ecb6f6d839ecc77fa10a2c573cc069990cc" \
            ["g3s"]="57569ecb6f6d839ecc77fa10a2c573cc069990cc" \
            ["g4dn"]="57569ecb6f6d839ecc77fa10a2c573cc069990cc" \
            ["p2"]="57569ecb6f6d839ecc77fa10a2c573cc069990cc" \
            ["p3"]="57569ecb6f6d839ecc77fa10a2c573cc069990cc" \
        )

        # CUDA Files that need to be installed ~1.4GB
        #   First one is main installation
        #   Subsequent files are patches which need to be applied in order
        #   Order in the arrays below matters
        # https://developer.nvidia.com/cuda-downloads
        cuda_files=( \
        "https://developer.nvidia.com/compute/cuda/9.1/Prod/local_installers/cuda_9.1.85_387.26_linux" \
        "https://developer.nvidia.com/compute/cuda/9.1/Prod/patches/1/cuda_9.1.85.1_linux" \
        "https://developer.nvidia.com/compute/cuda/9.1/Prod/patches/2/cuda_9.1.85.2_linux" \
        "https://developer.nvidia.com/compute/cuda/9.1/Prod/patches/3/cuda_9.1.85.3_linux" \
        )

        cuda_files_checksums=( \
        "1540658f4fe657dddd8b0899555b7468727d4aa8" \
        "7ec6970ecd81163b0d02ef30d35599e7fd6e97d8" \
        "cfa3b029b58fc117d8ce510a70efc848924dd565" \
        "6269a2c5784b08997edb97ea0020fb4e6c8769ed" \
        )
        ;;
    10.0)
        class_to_driver_file=( \
            ["g3"]="http://us.download.nvidia.com/tesla/410.129/NVIDIA-Linux-x86_64-410.129-diagnostic.run" \
            ["g3s"]="http://us.download.nvidia.com/tesla/410.129/NVIDIA-Linux-x86_64-410.129-diagnostic.run" \
            ["g4dn"]="http://us.download.nvidia.com/tesla/410.129/NVIDIA-Linux-x86_64-410.129-diagnostic.run" \
            ["p2"]="http://us.download.nvidia.com/tesla/410.129/NVIDIA-Linux-x86_64-410.129-diagnostic.run" \
            ["p3"]="http://us.download.nvidia.com/tesla/410.129/NVIDIA-Linux-x86_64-410.129-diagnostic.run" \
        )

        class_to_driver_checksum=( \
            ["g3"]="e5d234cc8acb35f425f60e1923e07e7e50272d9c" \
            ["g3s"]="e5d234cc8acb35f425f60e1923e07e7e50272d9c" \
            ["g4dn"]="e5d234cc8acb35f425f60e1923e07e7e50272d9c" \
            ["p2"]="e5d234cc8acb35f425f60e1923e07e7e50272d9c" \
            ["p3"]="e5d234cc8acb35f425f60e1923e07e7e50272d9c" \
        )

        cuda_files=( \
        "http://developer.download.nvidia.com/compute/cuda/10.1/Prod/local_installers/cuda_10.1.243_418.87.00_linux.run" \
        )

        cuda_files_checksums=( \
        "36706e2c0fb7efa14aea6a3c889271d97fd3575d" \
        ) 
        ;;
    *) echo "CUDA ${CUDA_VERSION} not supported by kops hook" && exit 1
esac

containsElement () { for e in "${@:2}"; do [[ "$e" = "$1" ]] && return 0; done; return 1; }

#################################################
# Ensure that we are on a proper AWS GPU Instance

apt-get -y update
apt-get -y --no-upgrade install curl jq

AWS_INSTANCE_TYPE=$(curl -m 2 -fsSL http://169.254.169.254/latest/dynamic/instance-identity/document | jq -r ".instanceType" || true) # eg: p2.micro
AWS_INSTANCE_CLASS=$(echo $AWS_INSTANCE_TYPE | cut -d . -f 1 || true) # e.g. p2

if [[ -z $AWS_INSTANCE_TYPE ]] || [[ -z $AWS_INSTANCE_CLASS ]]; then
  echo "This machine is not an AWS instance"
  echo "  Exiting without installing GPU drivers"
  exit 1
fi

classnames=${!class_to_driver_file[@]} # e.g. [ "g2", "g3", "g3s", "p2", "p3" ]
if ! containsElement $AWS_INSTANCE_CLASS $classnames; then
  echo "This machine is an AWS instance, but not a GPU instance"
  echo "  Exiting without installing GPU drivers"
  exit 1
fi

echo "Identified machine as AWS_INSTANCE_TYPE[$AWS_INSTANCE_TYPE] AWS_INSTANCE_CLASS[$AWS_INSTANCE_CLASS]"

#################################################
# Install dependencies

# Install GCC and linux headers on the host machine
#   The NVIDIA driver build must be compiled with the same version of GCC as
#   the kernel.  In addition, linux-headers are machine image specific.
#   Install with --no-upgrade so that the c-libs are not upgraded, possibly
#   breaking programs and requiring restart
apt-get -y update
apt-get -y --no-upgrade install gcc libc-dev libxml2-dev linux-headers-$(uname -r)
apt-get -y clean
apt-get -y autoremove

#################################################
# Unload open-source nouveau driver if it exists
#   The nvidia drivers won't install otherwise
#   "g3" instances in particular have this module auto-loaded
modprobe -r nouveau || true

#################################################
# Download and install the Nvidia drivers and cuda libraries

# Create list of URLs and Checksums by merging driver item with array of cuda files
downloads=(${class_to_driver_file[$AWS_INSTANCE_CLASS]} ${cuda_files[@]})
checksums=(${class_to_driver_checksum[$AWS_INSTANCE_CLASS]} ${cuda_files_checksums[@]})

# Download, verify, and execute each file
length=${#downloads[@]}
for (( i=0; i<${length}; i++ )); do
  download=${downloads[$i]}
  filename=$(basename $download)
  filepath_installed="${CACHE_DIR}/${filename}.installed"

  # Install the Nvidia driver and cuda libs
  if [[ -f $filepath_installed ]]; then
    echo "Detected prior install of file $filename on host"
  else
    checksum=${checksums[$i]}
    filepath="${CACHE_DIR}/${filename}"

    echo "Checking for file at $filepath"
    if [[ ! -f $filepath ]] || ! (echo "$checksum  $filepath" | sha1sum -c - 2>&1 >/dev/null); then
      echo "Downloading $download"
      curl -L $download > $filepath
      chmod a+x $filepath
    fi

    echo "Verifying sha1sum of file at $filepath"
    if ! (echo "$checksum  $filepath" | sha1sum -c -); then
      echo "Failed to verify sha1sum for file at $filepath"
      exit 1
    fi

    echo "Installing file $filename on host"
    if [[ $download =~ .*NVIDIA.* ]]; then
      # Install the nvidia package
      $filepath --accept-license --silent
      touch $filepath_installed # Mark successful installation
    elif [[ $download =~ .*local_installers.*cuda.* ]]; then
      # Install the primary cuda library
      $filepath --toolkit --silent
      touch $filepath_installed # Mark successful installation
    elif [[ $download =~ .*patches.*cuda.* ]]; then
      # Install an update to the primary cuda library
      $filepath --accept-eula --silent
      touch $filepath_installed # Mark successful installation
    else
      echo "Unable to handle file $filepath"
      exit 1
    fi
  fi
done

#################################################
# Output GPU info for debugging
nvidia-smi --list-gpus

#################################################
# Configure and Optimize Nvidia cards now that things are installed
#   AWS Optimizization Doc
#     https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/optimize_gpu.html
#   Nvidia Doc
#     http://developer.download.nvidia.com/compute/DCGM/docs/nvidia-smi-367.38.pdf

# Common configurations
nvidia-smi -pm 1
nvidia-smi --auto-boost-default=0
nvidia-smi --auto-boost-permission=0

# Custom configurations per class of nvidia video card
case "$AWS_INSTANCE_CLASS" in
"g2" | "g3" | "g3s")
  nvidia-smi -ac 2505,1177
  ;;
"p2")
  nvidia-smi -ac 2505,875
  nvidia-smi -acp 0
  ;;
"p3")
  nvidia-smi -ac 877,1530
  nvidia-smi -acp 0
  ;;
*)
  ;;
esac

#################################################
# Load the Kernel Module

if ! /sbin/modprobe nvidia-uvm; then
  echo "Unable to modprobe nvidia-uvm"
  exit 1
fi

# Ensure that the device node exists
if ! test -e /dev/nvidia-uvm; then
  # Find out the major device number used by the nvidia-uvm driver
  D=`grep nvidia-uvm /proc/devices | awk '{print $1}'`
  mknod -m 666 /dev/nvidia-uvm c $D 0
fi

###########################################################
# Restart Kubelet
#   Only necessary in the case of Accelerators (not Device Plugins)

echo "Restarting Kubelet"
systemctl restart kubelet.service
