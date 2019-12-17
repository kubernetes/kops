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

CACHE_DIR=/nvidia-device-plugin

# Load Passthrough environment variables from the original kops hook
source $CACHE_DIR/environment

# Support both deviceplugin and legacy (accelerator) GPU modes.
#   Default to 'deviceplugin' if env var is unset.
if [[ ! -v NVIDIA_DEVICE_PLUGIN_MODE ]]; then
    NVIDIA_DEVICE_PLUGIN_MODE='deviceplugin'
    echo "Defaulting to NVIDIA_DEVICE_PLUGIN_MODE='deviceplugin'"
fi

# Figure out which scripts should run
scripts=()
case "$NVIDIA_DEVICE_PLUGIN_MODE" in
    legacy)
        scripts+=("$CACHE_DIR/01-aws-nvidia-driver.sh")
        ;;
    deviceplugin)
        scripts+=("$CACHE_DIR/01-aws-nvidia-driver.sh")
        scripts+=("$CACHE_DIR/02-nvidia-docker.sh")
        ;;
    *)
	echo "Invalid NVIDIA_DEVICE_PLUGIN_MODE=$NVIDIA_DEVICE_PLUGIN_MODE"
	echo "  Valid values are 'deviceplugin' or 'legacy'"
	exit 1
esac

# Run the scripts
for script in "${scripts[@]}"; do
    echo "########## Starting $script ##########"
    $script 2>&1 | tee -a $CACHE_DIR/install.log
    echo "########## Finished $script ##########"
done
