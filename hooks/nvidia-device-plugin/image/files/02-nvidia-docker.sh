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
# Install nvidia-docker2

# This section is somewhat adapted from README at:
#   https://github.com/NVIDIA/nvidia-docker

#######################################
# Cleanup old nvidia-docker

# If you have nvidia-docker 1.0 installed: we need to remove it and all existing GPU containers
docker volume ls -q -f driver=nvidia-docker | xargs -r -I{} -n1 docker ps -q -a -f volume={} | xargs -r docker rm -f

# Remove the old nvidia-docker if it exists
apt-get purge -y nvidia-docker || true

#######################################
# Add package repositories

# Add the package repository for docker-ce
curl -fsSL https://download.docker.com/linux/debian/gpg | \
  apt-key add -
echo 'deb [arch=amd64] https://download.docker.com/linux/debian stretch stable' | \
  tee /etc/apt/sources.list.d/docker-ce.list

# Add the package repository for nvidia-docker
curl -s -L https://nvidia.github.io/nvidia-docker/gpgkey | \
  apt-key add -
distribution=$(. /etc/os-release;echo $ID$VERSION_ID)
curl -s -L https://nvidia.github.io/nvidia-docker/$distribution/nvidia-docker.list | \
  tee /etc/apt/sources.list.d/nvidia-docker.list

# Override the default runtime with the one from nvidia
#   Also explicity set the storage-driver to the prior 'overlay'
cat << 'EOF' > /etc/docker/daemon.json
{
    "default-runtime": "nvidia",
    "runtimes": {
        "nvidia": {
            "path": "/usr/bin/nvidia-container-runtime",
            "runtimeArgs": []
        }
    },
    "storage-driver": "overlay"
}
EOF

# Install nvidia-docker2 and reload the Docker daemon configuration
# Note that the nvidia-docker version must match the docker-ce version
# --force-confold prevents prompt for replacement of daemon.json
apt-get -y update
# Stop protokube to ensure not bring kubelet up again
systemctl stop protokube
# Stop kubelet to ensure not bring stopped containers up again and leak
# them as orphan containers
systemctl stop kubelet
# pin versions https://github.com/NVIDIA/nvidia-docker/wiki/Frequently-Asked-Questions#how-do-i-install-20-if-im-not-using-the-latest-docker-version
apt-get install -y --allow-downgrades -o Dpkg::Options::="--force-confold" \
  nvidia-docker2=2.0.3+docker18.09.4-1 \
  nvidia-container-runtime=2.0.0+docker18.09.4-1 \
  docker-ce=5:18.09.4~3-0~debian-stretch

# Disable a few things that break docker-ce/gpu support upon reboot:
#  Upon boot, the kops-configuration.service systemd unit sets up and starts
#  the cloud-init.service which runs nodeup which forces docker-ce to a
#  specific version that is a downgrade and incompatible with nvidia-docker2.
#  Permanently disable these systemd units via masking.
systemctl mask cloud-init.service
systemctl mask kops-configuration.service

# Restore protokube and protokube will bring up kubelet
systemctl start protokube
# Seems protokube won't bring up kubelet, so start kubelet separately
systemctl start kubelet
