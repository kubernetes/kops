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
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"mime/multipart"
	"net/textproto"
	"strings"
	"text/template"

	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/mirrors"
)

var nodeUpTemplate = `#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

NODEUP_URL_AMD64={{ NodeUpSourceAmd64 }}
NODEUP_HASH_AMD64={{ NodeUpSourceHashAmd64 }}
NODEUP_URL_ARM64={{ NodeUpSourceArm64 }}
NODEUP_HASH_ARM64={{ NodeUpSourceHashArm64 }}

{{ EnvironmentVariables }}

{{ ProxyEnv }}

{{ SetSysctls }}

function ensure-install-dir() {
  INSTALL_DIR="/opt/kops"
  # On ContainerOS, we install under /var/lib/toolbox; /opt is ro and noexec
  if [[ -d /var/lib/toolbox ]]; then
    INSTALL_DIR="/var/lib/toolbox/kops"
  fi
  mkdir -p ${INSTALL_DIR}/bin
  mkdir -p ${INSTALL_DIR}/conf
  cd ${INSTALL_DIR}
}

# Retry a download until we get it. args: name, sha, urls
download-or-bust() {
  local -r file="$1"
  local -r hash="$2"
  local -r urls=( $(split-commas "$3") )

  if [[ -f "${file}" ]]; then
    if ! validate-hash "${file}" "${hash}"; then
      rm -f "${file}"
    else
      return 0
    fi
  fi

  while true; do
    for url in "${urls[@]}"; do
      commands=(
        "curl -f --compressed -Lo "${file}" --connect-timeout 20 --retry 6 --retry-delay 10"
        "wget --compression=auto -O "${file}" --connect-timeout=20 --tries=6 --wait=10"
        "curl -f -Lo "${file}" --connect-timeout 20 --retry 6 --retry-delay 10"
        "wget -O "${file}" --connect-timeout=20 --tries=6 --wait=10"
      )
      for cmd in "${commands[@]}"; do
        echo "Attempting download with: ${cmd} {url}"
        if ! (${cmd} "${url}"); then
          echo "== Download failed with ${cmd} =="
          continue
        fi
        if ! validate-hash "${file}" "${hash}"; then
          echo "== Hash validation of ${url} failed. Retrying. =="
          rm -f "${file}"
        else
          echo "== Downloaded ${url} (SHA256 = ${hash}) =="
          return 0
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

function download-release() {
  case "$(uname -m)" in
  x86_64*|i?86_64*|amd64*)
    NODEUP_URL="${NODEUP_URL_AMD64}"
    NODEUP_HASH="${NODEUP_HASH_AMD64}"
    ;;
  aarch64*|arm64*)
    NODEUP_URL="${NODEUP_URL_ARM64}"
    NODEUP_HASH="${NODEUP_HASH_ARM64}"
    ;;
  *)
    echo "Unsupported host arch: $(uname -m)" >&2
    exit 1
    ;;
  esac

  cd ${INSTALL_DIR}/bin
  download-or-bust nodeup "${NODEUP_HASH}" "${NODEUP_URL}"

  chmod +x nodeup

  echo "Running nodeup"
  # We can't run in the foreground because of https://github.com/docker/docker/issues/23793
  ( cd ${INSTALL_DIR}/bin; ./nodeup --install-systemd-unit --conf=${INSTALL_DIR}/conf/kube_env.yaml --v=8  )
}

####################################################################################

if [[ ! -s /etc/machine-id ]]; then
  echo "Generating machine-id"
  /bin/systemd-machine-id-setup || echo "failed to set up ensure machine-id configured"
fi

echo "== nodeup node config starting =="
ensure-install-dir

{{ if CompressUserData -}}
echo "{{ GzipBase64 ClusterSpec }}" | base64 -d | gzip -d > conf/cluster_spec.yaml
{{- else -}}
cat > conf/cluster_spec.yaml << '__EOF_CLUSTER_SPEC'
{{ ClusterSpec }}
__EOF_CLUSTER_SPEC
{{- end }}

{{ if CompressUserData -}}
echo "{{ GzipBase64 KubeEnv }}" | base64 -d | gzip -d > conf/kube_env.yaml
{{- else -}}
cat > conf/kube_env.yaml << '__EOF_KUBE_ENV'
{{ KubeEnv }}
__EOF_KUBE_ENV
{{- end }}

download-release
echo "== nodeup node config done =="
`

// NodeUpScript is responsible for creating the nodeup script
type NodeUpScript struct {
	NodeUpAssets         map[architectures.Architecture]*mirrors.MirroredAsset
	KubeEnv              string
	CompressUserData     bool
	SetSysctls           string
	ProxyEnv             func() (string, error)
	EnvironmentVariables func() (string, error)
	ClusterSpec          func() (string, error)
}

func funcEmptyString() (string, error) {
	return "", nil
}

func (b *NodeUpScript) Build() (fi.Resource, error) {
	if b.ProxyEnv == nil {
		b.ProxyEnv = funcEmptyString
	}
	if b.EnvironmentVariables == nil {
		b.EnvironmentVariables = funcEmptyString
	}
	if b.ClusterSpec == nil {
		b.ClusterSpec = funcEmptyString
	}

	functions := template.FuncMap{
		"NodeUpSourceAmd64": func() string {
			if b.NodeUpAssets[architectures.ArchitectureAmd64] != nil {
				return strings.Join(b.NodeUpAssets[architectures.ArchitectureAmd64].Locations, ",")
			}
			return ""
		},
		"NodeUpSourceHashAmd64": func() string {
			if b.NodeUpAssets[architectures.ArchitectureAmd64] != nil {
				return b.NodeUpAssets[architectures.ArchitectureAmd64].Hash.Hex()
			}
			return ""
		},
		"NodeUpSourceArm64": func() string {
			if b.NodeUpAssets[architectures.ArchitectureArm64] != nil {
				return strings.Join(b.NodeUpAssets[architectures.ArchitectureArm64].Locations, ",")
			}
			return ""
		},
		"NodeUpSourceHashArm64": func() string {
			if b.NodeUpAssets[architectures.ArchitectureArm64] != nil {
				return b.NodeUpAssets[architectures.ArchitectureArm64].Hash.Hex()
			}
			return ""
		},

		"KubeEnv": func() string {
			return b.KubeEnv
		},

		"GzipBase64": func(data string) (string, error) {
			return gzipBase64(data)
		},

		"CompressUserData": func() bool {
			return b.CompressUserData
		},

		"SetSysctls": func() string {
			return b.SetSysctls
		},

		"ProxyEnv":             b.ProxyEnv,
		"EnvironmentVariables": b.EnvironmentVariables,
		"ClusterSpec":          b.ClusterSpec,
	}

	return newTemplateResource("nodeup", nodeUpTemplate, functions, nil)
}

func gzipBase64(data string) (string, error) {
	var b bytes.Buffer
	gz := gzip.NewWriter(&b)

	_, err := gz.Write([]byte(data))
	if err != nil {
		return "", err
	}

	if err = gz.Flush(); err != nil {
		return "", err
	}

	if err = gz.Close(); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(b.Bytes()), nil
}

// AWSMultipartMIME returns a MIME Multi Part Archive containing the nodeup (bootstrap) script
// and any additional User Data passed to using AdditionalUserData in the IG Spec
func AWSMultipartMIME(bootScript string, ig *kops.InstanceGroup) (string, error) {
	userData := bootScript

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
			err := writeUserDataPart(mimeWriter, "nodeup.sh", "text/x-shellscript", []byte(bootScript))
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

		userData = buffer.String()
	}

	return userData, nil
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
