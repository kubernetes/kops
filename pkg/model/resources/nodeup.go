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
	"os"
	"sort"
	"strconv"
	"strings"
	"text/template"

	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/cloudup/scaleway"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/architectures"
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
  echo "== Downloading $1 with hash $2 from $3 =="
  local -r file="$1"
  local -r hash="$2"
  local -a urls
  mapfile -t urls < <(split-commas "$3")

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
        "curl -f --compressed -Lo ${file} --connect-timeout 20 --retry 6 --retry-delay 10"
        "wget --compression=auto -O ${file} --connect-timeout=20 --tries=6 --wait=10"
        "curl -f -Lo ${file} --connect-timeout 20 --retry 6 --retry-delay 10"
        "wget -O ${file} --connect-timeout=20 --tries=6 --wait=10"
      )
      for cmd in "${commands[@]}"; do
        echo "== Downloading ${url} using ${cmd} =="
        if ! (${cmd} "${url}"); then
          echo "== Failed to download ${url} using ${cmd} =="
          continue
        fi
        if ! validate-hash "${file}" "${hash}"; then
          echo "== Failed to validate hash for ${url} =="
          rm -f "${file}"
        else
          echo "== Downloaded ${url} with hash ${hash} =="
          return 0
        fi
      done
    done

    echo "== All downloads failed; sleeping before retrying =="
    sleep 60
  done
}

validate-hash() {
  local -r file="$1"
  local -r expected="$2"
  local actual

  actual=$(sha256sum "${file}" | awk '{ print $1 }') || true
  if [[ "${actual}" != "${expected}" ]]; then
    echo "== File ${file} is corrupted; hash ${actual} doesn't match expected ${expected} =="
    return 1
  fi
}

function split-commas() {
  echo "$1" | tr "," "\n"
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

  echo "== Running nodeup =="
  # We can't run in the foreground because of https://github.com/docker/docker/issues/23793
  ( cd ${INSTALL_DIR}/bin; ./nodeup --install-systemd-unit --conf=${INSTALL_DIR}/conf/kube_env.yaml --v=8  )
}

####################################################################################

/bin/systemd-machine-id-setup || echo "== Failed to initialize the machine ID; ensure machine-id configured =="

{{- if eq GetCloudProvider "digitalocean" }}
  # DO has machine-id baked into the image and journald should be flushed
  # to use the new machine-id
  systemctl restart systemd-journald
{{- end }}

echo "== nodeup node config starting =="
ensure-install-dir

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
	NodeUpAssets         map[architectures.Architecture]*assets.MirroredAsset
	BootConfig           *nodeup.BootConfig
	CompressUserData     bool
	SetSysctls           string
	CloudProvider        string
	ProxyEnv             func() (string, error)
	EnvironmentVariables func() (string, error)
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

		"KubeEnv": func() (string, error) {
			bootConfigData, err := utils.YamlMarshal(b.BootConfig)
			if err != nil {
				return "", fmt.Errorf("error converting boot config to yaml: %w", err)
			}

			return string(bootConfigData), nil
		},

		"GzipBase64": func(data string) (string, error) {
			return gzipBase64(data)
		},

		"CompressUserData": func() bool {
			return b.CompressUserData
		},

		"GetCloudProvider": func() string {
			return b.CloudProvider
		},

		"SetSysctls": func() string {
			return b.SetSysctls
		},

		"ProxyEnv":             b.ProxyEnv,
		"EnvironmentVariables": b.EnvironmentVariables,
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

func buildEnvironmentVariables(cluster *kops.Cluster, ig *kops.InstanceGroup) (map[string]string, error) {
	env := make(map[string]string)

	if os.Getenv("GOSSIP_DNS_CONN_LIMIT") != "" {
		env["GOSSIP_DNS_CONN_LIMIT"] = os.Getenv("GOSSIP_DNS_CONN_LIMIT")
	}

	if os.Getenv("S3_ENDPOINT") != "" {
		if ig.IsControlPlane() {
			env["S3_ENDPOINT"] = os.Getenv("S3_ENDPOINT")
			env["S3_REGION"] = os.Getenv("S3_REGION")
			env["S3_ACCESS_KEY_ID"] = os.Getenv("S3_ACCESS_KEY_ID")
			env["S3_SECRET_ACCESS_KEY"] = os.Getenv("S3_SECRET_ACCESS_KEY")
		}
	}

	if cluster.GetCloudProvider() == kops.CloudProviderOpenstack {

		osEnvs := []string{
			"OS_TENANT_ID", "OS_TENANT_NAME", "OS_PROJECT_ID", "OS_PROJECT_NAME",
			"OS_PROJECT_DOMAIN_NAME", "OS_PROJECT_DOMAIN_ID",
			"OS_DOMAIN_NAME", "OS_DOMAIN_ID",
			"OS_AUTH_URL",
			"OS_REGION_NAME",
		}

		appCreds := os.Getenv("OS_APPLICATION_CREDENTIAL_ID") != "" && os.Getenv("OS_APPLICATION_CREDENTIAL_SECRET") != ""
		if appCreds {
			osEnvs = append(osEnvs,
				"OS_APPLICATION_CREDENTIAL_ID",
				"OS_APPLICATION_CREDENTIAL_SECRET",
			)
		} else {
			klog.Warning("exporting username and password. Consider using application credentials instead.")
			osEnvs = append(osEnvs,
				"OS_USERNAME",
				"OS_PASSWORD",
			)
		}

		// credentials needed always in control-plane and when using gossip also in nodes
		passEnvs := false
		if ig.IsControlPlane() || cluster.UsesLegacyGossip() {
			passEnvs = true
		}
		// Pass in required credentials when using user-defined swift endpoint
		if os.Getenv("OS_AUTH_URL") != "" && passEnvs {
			for _, envVar := range osEnvs {
				env[envVar] = fmt.Sprintf("'%s'", os.Getenv(envVar))
			}
		}
	}

	if cluster.GetCloudProvider() == kops.CloudProviderDO {
		if ig.IsControlPlane() {
			doToken := os.Getenv("DIGITALOCEAN_ACCESS_TOKEN")
			if doToken != "" {
				env["DIGITALOCEAN_ACCESS_TOKEN"] = doToken
			}
		}
	}

	if cluster.GetCloudProvider() == kops.CloudProviderHetzner && (ig.IsControlPlane() || cluster.UsesLegacyGossip()) {
		hcloudToken := os.Getenv("HCLOUD_TOKEN")
		if hcloudToken != "" {
			env["HCLOUD_TOKEN"] = hcloudToken
		}
	}

	if cluster.GetCloudProvider() == kops.CloudProviderAWS {
		region, err := awsup.FindRegion(cluster)
		if err != nil {
			return nil, err
		}
		if region == "" {
			klog.Warningf("unable to determine cluster region")
		} else {
			env["AWS_REGION"] = region
		}
	}

	if cluster.GetCloudProvider() == kops.CloudProviderAzure {
		env["AZURE_STORAGE_ACCOUNT"] = os.Getenv("AZURE_STORAGE_ACCOUNT")
		azureEnv := os.Getenv("AZURE_ENVIRONMENT")
		if azureEnv != "" {
			env["AZURE_ENVIRONMENT"] = os.Getenv("AZURE_ENVIRONMENT")
		}
	}

	if cluster.GetCloudProvider() == kops.CloudProviderScaleway && (ig.IsControlPlane() || cluster.UsesLegacyGossip()) {
		profile, err := scaleway.CreateValidScalewayProfile()
		if err != nil {
			return nil, err
		}
		env["SCW_ACCESS_KEY"] = fi.ValueOf(profile.AccessKey)
		env["SCW_SECRET_KEY"] = fi.ValueOf(profile.SecretKey)
		env["SCW_DEFAULT_PROJECT_ID"] = fi.ValueOf(profile.DefaultProjectID)
	}

	return env, nil
}

func (b *NodeUpScript) WithEnvironmentVariables(cluster *kops.Cluster, ig *kops.InstanceGroup) {
	b.EnvironmentVariables = func() (string, error) {
		env, err := buildEnvironmentVariables(cluster, ig)
		if err != nil {
			return "", err
		}

		// Sort keys to have a stable sequence of "export xx=xxx"" statements
		var keys []string
		for k := range env {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		var b bytes.Buffer
		for _, k := range keys {
			b.WriteString(fmt.Sprintf("export %s=%s\n", k, env[k]))
		}
		return b.String(), nil
	}

}

func createProxyEnv(ps *kops.EgressProxySpec) (string, error) {
	var buffer bytes.Buffer

	if ps != nil && ps.HTTPProxy.Host != "" {
		var httpProxyURL string

		// TODO double check that all the code does this
		// TODO move this into a validate so we can enforce the string syntax
		if !strings.HasPrefix(ps.HTTPProxy.Host, "http://") {
			httpProxyURL = "http://"
		}

		if ps.HTTPProxy.Port != 0 {
			httpProxyURL += ps.HTTPProxy.Host + ":" + strconv.Itoa(ps.HTTPProxy.Port)
		} else {
			httpProxyURL += ps.HTTPProxy.Host
		}

		// Set env variables for base environment
		buffer.WriteString(`{` + "\n")
		buffer.WriteString(`  echo "http_proxy=` + httpProxyURL + `"` + "\n")
		buffer.WriteString(`  echo "https_proxy=` + httpProxyURL + `"` + "\n")
		buffer.WriteString(`  echo "no_proxy=` + ps.ProxyExcludes + `"` + "\n")
		buffer.WriteString(`  echo "NO_PROXY=` + ps.ProxyExcludes + `"` + "\n")
		buffer.WriteString(`} >> /etc/environment` + "\n")

		// Load the proxy environment variables
		buffer.WriteString("while read -r in; do export \"${in?}\"; done < /etc/environment\n")

		// Set env variables for package manager depending on OS Distribution (N/A for Flatcar)
		// Note: Nodeup will source the `/etc/environment` file within docker config in the correct location
		buffer.WriteString("case $(cat /proc/version) in\n")
		buffer.WriteString("*[Dd]ebian* | *[Uu]buntu*)\n")
		buffer.WriteString(`  echo "Acquire::http::Proxy \"` + httpProxyURL + `\";" > /etc/apt/apt.conf.d/30proxy ;;` + "\n")
		buffer.WriteString("*[Rr]ed[Hh]at*)\n")
		buffer.WriteString(`  echo "proxy=` + httpProxyURL + `" >> /etc/yum.conf ;;` + "\n")
		buffer.WriteString("esac\n")

		// Set env variables for systemd
		buffer.WriteString(`echo "DefaultEnvironment=\"http_proxy=` + httpProxyURL + `\" \"https_proxy=` + httpProxyURL + `\"`)
		buffer.WriteString(` \"NO_PROXY=` + ps.ProxyExcludes + `\" \"no_proxy=` + ps.ProxyExcludes + `\""`)
		buffer.WriteString(" >> /etc/systemd/system.conf\n")

		// Restart stuff
		buffer.WriteString("systemctl daemon-reload\n")
		buffer.WriteString("systemctl daemon-reexec\n")
	}
	return buffer.String(), nil
}

func (b *NodeUpScript) WithProxyEnv(cluster *kops.Cluster) {
	b.ProxyEnv = func() (string, error) {
		return createProxyEnv(cluster.Spec.Networking.EgressProxy)
	}
}

// By setting some sysctls early, we avoid broken configurations that prevent nodeup download.
// See https://github.com/kubernetes/kops/issues/10206 for details.
func (s *NodeUpScript) WithSysctls() {
	var b bytes.Buffer

	// Based on https://github.com/kubernetes/kops/issues/10206#issuecomment-766852332
	b.WriteString("sysctl -w net.core.rmem_max=16777216 || true\n")
	b.WriteString("sysctl -w net.core.wmem_max=16777216 || true\n")
	b.WriteString("sysctl -w net.ipv4.tcp_rmem='4096 87380 16777216' || true\n")
	b.WriteString("sysctl -w net.ipv4.tcp_wmem='4096 87380 16777216' || true\n")

	s.SetSysctls = b.String()
}
