/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vspheretasks

// Template for user-data file in the cloud-init ISO
const userDataTemplate = `#cloud-config
write_files:
  - content: |
$SCRIPT
    owner: root:root
    path: /root/script.sh
    permissions: "0644"
  - content: |
$DNS_SCRIPT
    owner: root:root
    path: /root/update_dns.sh
    permissions: "0644"
  - content: |
$VM_UUID
    owner: root:root
    path: /etc/vmware/vm_uuid
    permissions: "0644"
  - content: |
$VOLUME_SCRIPT
    owner: root:root
    path: /vol-metadata/metadata.json
    permissions: "0644"

runcmd:
  - bash /root/update_dns.sh 2>&1 > /var/log/update_dns.log
  - bash /root/script.sh 2>&1 > /var/log/script.log`

// Template for meta-data file in the cloud-init ISO
const metaDataTemplate = `instance-id: $INSTANCE_ID
local-hostname: $LOCAL_HOST_NAME`
