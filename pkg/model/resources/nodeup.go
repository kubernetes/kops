/*
Copyright 2019 The Kubernetes Authors.

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

package resources

import (
	"bufio"
	"bytes"
	"fmt"
	"mime/multipart"
	"net/textproto"

	"k8s.io/kops/pkg/apis/kops"
)

var NodeUpTemplate = `#!/bin/bash
# Copyright 2016 The Kubernetes Authors All rights reserved.
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

set -o errexit
set -o nounset
set -o pipefail

NODEUP_URL={{ NodeUpSource }}
NODEUP_HASH={{ NodeUpSourceHash }}

{{ EnvironmentVariables }}

{{ ProxyEnv }}

function ensure-install-dir() {
  INSTALL_DIR="/var/cache/kubernetes-install"
  # On ContainerOS, we install to /var/lib/toolbox install (because of noexec)
  if [[ -d /var/lib/toolbox ]]; then
    INSTALL_DIR="/var/lib/toolbox/kubernetes-install"
  fi
  mkdir -p ${INSTALL_DIR}
  cd ${INSTALL_DIR}
}

# Retry a download until we get it. args: name, sha, url1, url2...
download-or-bust() {
  local -r file="$1"
  local -r hash="$2"
  shift 2

  urls=( $* )
  while true; do
    for url in "${urls[@]}"; do
      commands=(
        "curl -f --ipv4 --compressed -Lo "${file}" --connect-timeout 20 --retry 6 --retry-delay 10"
        "wget --inet4-only --compression=auto -O "${file}" --connect-timeout=20 --tries=6 --wait=10"
        "curl -f --ipv4 -Lo "${file}" --connect-timeout 20 --retry 6 --retry-delay 10"
        "wget --inet4-only -O "${file}" --connect-timeout=20 --tries=6 --wait=10"
      )
      for cmd in "${commands[@]}"; do
        echo "Attempting download with: ${cmd} {url}"
        if ! (${cmd} "${url}"); then
          echo "== Download failed with ${cmd} =="
          continue
        fi
        if [[ -n "${hash}" ]] && ! validate-hash "${file}" "${hash}"; then
          echo "== Hash validation of ${url} failed. Retrying. =="
          rm -f "${file}"
        else
          if [[ -n "${hash}" ]]; then
            echo "== Downloaded ${url} (SHA1 = ${hash}) =="
          else
            echo "== Downloaded ${url} =="
          fi
          return
        fi
      done
    done

    echo "All downloads failed; sleeping before retrying"
    sleep 60
  done
}

validate-hash() {
  local -r file="$1"
  local -r expected="$2"
  local actual

  actual=$(sha256sum ${file} | awk '{ print $1 }') || true
  if [[ "${actual}" != "${expected}" ]]; then
    echo "== ${file} corrupted, hash ${actual} doesn't match expected ${expected} =="
    return 1
  fi
}

function split-commas() {
  echo $1 | tr "," "\n"
}

function try-download-release() {
  # TODO(zmerlynn): Now we REALLY have no excuse not to do the reboot
  # optimization.

  local -r nodeup_urls=( $(split-commas "${NODEUP_URL}") )
  if [[ -n "${NODEUP_HASH:-}" ]]; then
    local -r nodeup_hash="${NODEUP_HASH}"
  else
  # TODO: Remove?
    echo "Downloading sha256 (not found in env)"
    download-or-bust nodeup.sha256 "" "${nodeup_urls[@]/%/.sha256}"
    local -r nodeup_hash=$(cat nodeup.sha256)
  fi

  echo "Downloading nodeup (${nodeup_urls[@]})"
  download-or-bust nodeup "${nodeup_hash}" "${nodeup_urls[@]}"

  chmod +x nodeup
}

function download-release() {
  # In case of failure checking integrity of release, retry.
  until try-download-release; do
    sleep 15
    echo "Couldn't download release. Retrying..."
  done

  echo "Running nodeup"
  # We can't run in the foreground because of https://github.com/docker/docker/issues/23793
  ( cd ${INSTALL_DIR}; ./nodeup --install-systemd-unit --conf=${INSTALL_DIR}/kube_env.yaml --v=8  )
}

function run-bootstrap-scripts() {
	cat ./bootstrap_scripts.txt | while read scripturl scripthash; do
		try-run-bootstrap-script "$scripthash" "$scripturl"
	done
}

function try-run-bootstrap-script() {
	local -r scripthash="$1"
	local -r scripturls=( $(split-commas "$2") )
	local -r scriptname="${scripturls[0]##*/}"

	if [[ -n "${scripthash}" ]]; then
		local -r scriptsha1="${scripthash}"
	else
		echo "Downloading ${scriptname}.sha1 (not found in env)"
		download-or-bust "" "${scripturls[@]/%/.sha1}"
		local -r scriptsha1=$(cat "${scriptname}.sha1")
	fi

	echo "Downloading ${scriptname} (${scripturls[@]})"
	download-or-bust "${scriptsha1}" "${scripturls[@]}"

	echo "=== Running ${scriptname} ==="
	chmod +x "$scriptname"
	./$scriptname
	echo "=== Completed ${scriptname} ==="
}

####################################################################################

/bin/systemd-machine-id-setup || echo "failed to set up ensure machine-id configured"

echo "== nodeup node config starting =="
ensure-install-dir

cat > bootstrap_scripts.txt << '__EOF_BOOTSTRAP_SCRIPTS'
{{ BootstrapHooks }}
__EOF_BOOTSTRAP_SCRIPTS

cat > cluster_spec.yaml << '__EOF_CLUSTER_SPEC'
{{ ClusterSpec }}
__EOF_CLUSTER_SPEC

cat > ig_spec.yaml << '__EOF_IG_SPEC'
{{ IGSpec }}
__EOF_IG_SPEC

cat > kube_env.yaml << '__EOF_KUBE_ENV'
{{ KubeEnv }}
__EOF_KUBE_ENV

run-bootstrap-scripts
download-release
echo "== nodeup node config done =="
`

// AWSNodeUpTemplate returns a MIME Multi Part Archive containing the nodeup (bootstrap) script
// and any additional User Data passed to using AdditionalUserData in the IG Spec
func AWSNodeUpTemplate(ig *kops.InstanceGroup) (string, error) {

	userDataTemplate := NodeUpTemplate

	if len(ig.Spec.AdditionalUserData) > 0 {
		/* Create a buffer to hold the user-data*/
		buffer := bytes.NewBufferString("")
		writer := bufio.NewWriter(buffer)

		mimeWriter := multipart.NewWriter(writer)

		// we explicitly set the boundary to make testing easier.
		boundary := "MIMEBOUNDARY"
		if err := mimeWriter.SetBoundary(boundary); err != nil {
			return "", err
		}

		writer.Write([]byte(fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary)))
		writer.Write([]byte("MIME-Version: 1.0\r\n\r\n"))

		var err error
		if !ig.IsBastion() {
			err := writeUserDataPart(mimeWriter, "nodeup.sh", "text/x-shellscript", []byte(userDataTemplate))
			if err != nil {
				return "", err
			}
		}

		for _, d := range ig.Spec.AdditionalUserData {
			err = writeUserDataPart(mimeWriter, d.Name, d.Type, []byte(d.Content))
			if err != nil {
				return "", err
			}
		}

		writer.Write([]byte(fmt.Sprintf("\r\n--%s--\r\n", boundary)))

		writer.Flush()
		mimeWriter.Close()

		userDataTemplate = buffer.String()
	}

	return userDataTemplate, nil

}

func writeUserDataPart(mimeWriter *multipart.Writer, fileName string, contentType string, content []byte) error {
	header := textproto.MIMEHeader{}

	header.Set("Content-Type", contentType)
	header.Set("MIME-Version", "1.0")
	header.Set("Content-Transfer-Encoding", "7bit")
	header.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, fileName))

	partWriter, err := mimeWriter.CreatePart(header)
	if err != nil {
		return err
	}

	_, err = partWriter.Write(content)
	if err != nil {
		return err
	}

	return nil
}
