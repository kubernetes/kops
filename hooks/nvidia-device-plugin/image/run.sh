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

# Copy the setup scripts to the host
#   The kops hook automatically mounts the host root filesystem into the
#   container /rootfs
mkdir -p /rootfs/nvidia-device-plugin
cp -r /nvidia-device-plugin/* /rootfs/nvidia-device-plugin

# Setup the host systemd to run the systemd unit that runs setup scripts
ln -sf /nvidia-device-plugin/nvidia-device-plugin.service /rootfs/etc/systemd/system/nvidia-device-plugin.service

# Save the environment to be passed on to the systemd unit
(env | grep NVIDIA_DEVICE_PLUGIN > /rootfs/nvidia-device-plugin/environment) || true

# Kickoff host systemd unit that runs the setup scripts
#   'systemctl' within this docker container uses the mounted /run/systemd/*
#   volume from the host to control systemd on the host.
systemctl daemon-reload
systemctl start --no-block nvidia-device-plugin.service

exit 0
