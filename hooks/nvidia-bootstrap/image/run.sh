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

#!/bin/sh

# Simple early detection of nvidia card
grep 10de102d /proc/bus/pci/devices || exit 0

# p2.xlarge
# 00f0	10de102d	4b	        84000000	      100000000c	               0	        8200000c	               0	               0	               0	         1000000	       400000000	               0	         2000000	               0	               0	               0	nvidia


# This is pretty annoying.... note this is installed onto the host
chroot /rootfs apt-get update
chroot /rootfs apt-get install --yes gcc

mkdir -p /rootfs/tmp
cd /rootfs/tmp
# TODO: We can't download over SSL - presents an akamai cert
wget http://us.download.nvidia.com/XFree86/Linux-x86_64/375.39/NVIDIA-Linux-x86_64-375.39.run
echo '5e5b9fbf12f4f926ed70c1fe39f71d9d9f154abea0268b1cf035982b34bd7c94baef7667e4f647cc19a62702b46f63b3c3df9f1589261f7138ed2ff151af63cc  NVIDIA-Linux-x86_64-375.39.run' | sha3sum -c - || exit 1
chmod +x NVIDIA-Linux-x86_64-375.39.run
chroot /rootfs /tmp/NVIDIA-Linux-x86_64-375.39.run --accept-license --ui=none

cd /rootfs/tmp
wget https://developer.nvidia.com/compute/cuda/8.0/Prod2/local_installers/cuda_8.0.61_375.26_linux-run
chmod +x cuda_8.0.61_375.26_linux-run
# If we want to install samples as well, add: --samples
chroot /rootfs /tmp/cuda_8.0.61_375.26_linux-run --toolkit --silent

chroot /rootfs nvidia-smi -pm 1
chroot /rootfs nvidia-smi -acp 0
chroot /rootfs nvidia-smi --auto-boost-default=0
chroot /rootfs nvidia-smi --auto-boost-permission=0
chroot /rootfs nvidia-smi -ac 2505,875


# TODO: Problem ... why is this needed - why didn't this happen when we installed nvidia-uvm?
# TODO: Problem ... we need to restart kubelet

chroot /rootfs /sbin/modprobe nvidia-uvm

if [ "$?" -eq 0 ]; then
  # Find out the major device number used by the nvidia-uvm driver
  D=`grep nvidia-uvm /proc/devices | awk '{print $1}'`

  chroot /rootfs mknod -m 666 /dev/nvidia-uvm c $D 0
else
  echo "Unable to modprobe nvidia-uvm"
fi
