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
	"context"
	"encoding/base64"
	"fmt"
	"maps"
	"mime/multipart"
	"net/textproto"
	"net/url"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/smithy-go/encoding/httpbinding"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/apis/kops"
	"k8s.io/kops/pkg/apis/nodeup"
	"k8s.io/kops/pkg/assets"
	"k8s.io/kops/third_party/forked/text/template"
	"k8s.io/kops/upup/pkg/fi"
	"k8s.io/kops/upup/pkg/fi/cloudup/awsup"
	"k8s.io/kops/upup/pkg/fi/utils"
	"k8s.io/kops/util/pkg/architectures"
	"k8s.io/kops/util/pkg/vfs"
	"k8s.io/kops/util/pkg/vfs/openstackconfig"
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

{{- if UseS3Download }}
# Fetch a path from the instance metadata service. args: token, path
imds-get() {
  curl -s -f --noproxy '*' --connect-timeout 2 --max-time 5 \
    -H "X-aws-ec2-metadata-token: $1" "http://169.254.169.254/latest/$2"
}
{{- end }}

{{- if or UseGCSDownload UseS3Download UseBlobDownload }}

# Extract a string field from a JSON object. args: json, field
json-field() {
  local pattern="\"$2\"[[:space:]]*:[[:space:]]*\"([^\"]+)\""
  [[ $1 =~ $pattern ]] && printf '%s' "${BASH_REMATCH[1]}"
}
{{- end }}

# Retry a download until we get it. sha covers the uncompressed binary. args: name, sha, urls
download-or-bust() {
  echo "== Downloading $1 with hash $2 from $3 =="
  local -r file="$1"
  local -r hash="$2"
  local -a urls
  IFS=, read -r -a urls <<< "$3"

  if [[ -f "${file}" ]]; then
    if ! validate-hash "${file}" "${hash}"; then
      rm -f "${file}"
    else
      return 0
    fi
  fi

  while true; do
    for url in "${urls[@]}"; do
      echo "== Downloading ${url}.xz =="
{{- if UseGCSDownload }}
      local response token
      # Use the IP of the metadata server, to not depend on DNS this early in boot
      if ! response=$(curl -s -f --noproxy '*' --connect-timeout 2 --max-time 5 -H 'Metadata-Flavor: Google' "http://169.254.169.254/computeMetadata/v1/instance/service-accounts/default/token"); then
        echo "== Failed to get a service account token =="
      elif ! token=$(json-field "${response}" access_token); then
        echo "== Failed to parse the service account token =="
      # Pass the token through stdin so it does not appear in files, logs, or process arguments.
      elif ! echo "Authorization: Bearer ${token}" | curl -f -Lo "${file}.xz" --connect-timeout 20 --retry 6 --retry-delay 10 -H @- "https://storage.googleapis.com/${url#gs://}.xz"; then
        echo "== Failed to download ${url}.xz =="
        rm -f "${file}.xz"
{{- else if UseBlobDownload }}
      local rest account response token
      rest="${url#azureblob://}"
      account="${rest%%/*}"
      rest="${rest#*/}"
      # Use the IP of the metadata server, to not depend on DNS this early in boot.
      # The token request is brokered to Entra ID, so allow more time than for other clouds.
      if ! response=$(curl -s -f --noproxy '*' --connect-timeout 5 --max-time 30 -H 'Metadata: true' "http://169.254.169.254/metadata/identity/oauth2/token?api-version=2018-02-01&resource=https%3A%2F%2Fstorage.azure.com%2F"); then
        echo "== Failed to get a managed identity token =="
      elif ! token=$(json-field "${response}" access_token); then
        echo "== Failed to parse the managed identity token =="
      # Pass the token through stdin so it does not appear in files, logs, or process arguments.
      elif ! printf 'Authorization: Bearer %s\nx-ms-version: 2017-11-09\n' "${token}" |
        curl -f -Lo "${file}.xz" --connect-timeout 20 --retry 6 --retry-delay 10 -H @- \
          "https://${account}.blob.core.windows.net/${rest}.xz"; then
        echo "== Failed to download ${url}.xz =="
        rm -f "${file}.xz"
{{- else if UseS3Download }}
      local imds_token profile creds access_key secret_key session_token
      # Use the IP of the metadata server, to not depend on DNS this early in boot
      if ! imds_token=$(curl -s -f -X PUT --noproxy '*' --connect-timeout 2 --max-time 5 -H 'X-aws-ec2-metadata-token-ttl-seconds: 60' "http://169.254.169.254/latest/api/token"); then
        echo "== Failed to get an IMDS token =="
      elif ! profile=$(imds-get "${imds_token}" meta-data/iam/security-credentials/ | head -n 1); then
        echo "== Failed to get the instance profile name =="
      elif ! creds=$(imds-get "${imds_token}" "meta-data/iam/security-credentials/${profile}"); then
        echo "== Failed to get the instance profile credentials =="
      elif ! access_key=$(json-field "${creds}" AccessKeyId) || ! secret_key=$(json-field "${creds}" SecretAccessKey) || ! session_token=$(json-field "${creds}" Token); then
        echo "== Failed to parse the instance profile credentials =="
      # Pass credentials through stdin so they do not appear in files, logs, or process arguments.
      elif ! printf 'user "%s:%s"\nheader "x-amz-security-token: %s"\n' "${access_key}" "${secret_key}" "${session_token}" |
        curl -f -Lo "${file}.xz" --connect-timeout 20 --retry 6 --retry-delay 10 \
          --config - --aws-sigv4 "aws:amz:{{ S3Region }}:s3" \
          "https://s3.{{ S3Region }}.amazonaws.com/${url#s3://}.xz"; then
        echo "== Failed to download ${url}.xz =="
        rm -f "${file}.xz"
{{- else }}
      if ! curl -f -Lo "${file}.xz" --connect-timeout 20 --retry 6 --retry-delay 10 "${url}.xz"; then
        echo "== Failed to download ${url}.xz =="
        rm -f "${file}.xz"
{{- end }}
      elif ! xz -d "${file}.xz"; then
        echo "== Failed to decompress ${url}.xz =="
        rm -f "${file}" "${file}.xz"
      elif ! validate-hash "${file}" "${hash}"; then
        echo "== Failed to validate decompressed hash for ${url}.xz =="
        rm -f "${file}"
      else
        echo "== Downloaded ${url}.xz and validated decompressed hash ${hash} =="
        return 0
      fi
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

	// S3Region is baked into the script because SigV4 requires the bucket's actual region, which
	// an s3:// URL does not include.
	S3Region string
}

func funcEmptyString() (string, error) {
	return "", nil
}

func (b *NodeUpScript) nodeUpSource(arch architectures.Architecture) (string, error) {
	asset := b.NodeUpAssets[arch]
	if asset == nil {
		return "", nil
	}

	locations := slices.Clone(asset.Locations)
	for i, location := range locations {
		var escape func(string) (string, error)
		switch {
		case strings.HasPrefix(location, "s3://"):
			escape = escapeS3Location
		case strings.HasPrefix(location, "azureblob://"):
			escape = escapeBlobLocation
		default:
			continue
		}
		escaped, err := escape(location)
		if err != nil {
			return "", fmt.Errorf("escaping nodeup source %q: %w", location, err)
		}
		locations[i] = escaped
	}
	return strings.Join(locations, ","), nil
}

func escapeS3Location(location string) (string, error) {
	u, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parsing S3 location: %w", err)
	}
	if u.Scheme != "s3" || u.Host == "" {
		return "", fmt.Errorf("invalid S3 location")
	}

	return "s3://" + u.Host + httpbinding.EscapePath(u.Path, false), nil
}

func escapeBlobLocation(location string) (string, error) {
	u, err := url.Parse(location)
	if err != nil {
		return "", fmt.Errorf("parsing Azure Blob location: %w", err)
	}
	container, key, _ := strings.Cut(strings.TrimPrefix(u.Path, "/"), "/")
	// Reject ports, IPv6 hosts, userinfo, queries, and fragments, which the account-based
	// blob.core.windows.net URL cannot represent, so they fail here instead of in the boot retry loop.
	if u.Scheme != "azureblob" || u.Host == "" || u.Hostname() != u.Host || u.User != nil || u.RawQuery != "" || u.Fragment != "" || container == "" || key == "" {
		return "", fmt.Errorf("invalid Azure Blob location; expected azureblob://<account>/<container>/<key>")
	}

	return "azureblob://" + u.Host + httpbinding.EscapePath(u.Path, false), nil
}

func (b *NodeUpScript) Build() (fi.Resource, error) {
	if b.ProxyEnv == nil {
		b.ProxyEnv = funcEmptyString
	}
	if b.EnvironmentVariables == nil {
		b.EnvironmentVariables = funcEmptyString
	}

	if b.useS3Download() && b.S3Region == "" {
		return nil, fmt.Errorf("ResolveS3Region must be called before building a nodeup script with an s3:// source")
	}

	if b.useBlobDownload() {
		// The script hard-codes the public cloud blob.core.windows.net endpoint suffix.
		// Azure environment names are case-insensitive; AzureCloud is the CLI name of the public cloud.
		if azureEnv := os.Getenv("AZURE_ENVIRONMENT"); azureEnv != "" && !strings.EqualFold(azureEnv, "AzurePublicCloud") && !strings.EqualFold(azureEnv, "AzureCloud") {
			return nil, fmt.Errorf("downloading nodeup from an azureblob:// URL is not supported in Azure environment %q", azureEnv)
		}
	}

	functions := template.FuncMap{
		"NodeUpSourceAmd64": func() (string, error) {
			return b.nodeUpSource(architectures.ArchitectureAmd64)
		},
		"NodeUpSourceHashAmd64": func() string {
			if b.NodeUpAssets[architectures.ArchitectureAmd64] != nil {
				return b.NodeUpAssets[architectures.ArchitectureAmd64].Hash.Hex()
			}
			return ""
		},
		"NodeUpSourceArm64": func() (string, error) {
			return b.nodeUpSource(architectures.ArchitectureArm64)
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

		"GzipBase64": gzipBase64,

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

		"UseGCSDownload":  b.useGCSDownload,
		"UseS3Download":   b.useS3Download,
		"UseBlobDownload": b.useBlobDownload,
		"S3Region":        func() string { return b.S3Region },
	}

	return newTemplateResource("nodeup", nodeUpTemplate, functions, nil)
}

// useGCSDownload reports whether nodeup is downloaded from a GCS bucket, which has to be done
// with the credentials of the instance service account.
func (b *NodeUpScript) useGCSDownload() bool {
	return b.firstLocationWithScheme("gs://") != ""
}

func (b *NodeUpScript) firstLocationWithScheme(scheme string) string {
	// Sort architectures so region selection is deterministic and independent of KOPS_ARCH.
	for _, arch := range slices.Sorted(maps.Keys(b.NodeUpAssets)) {
		asset := b.NodeUpAssets[arch]
		if asset == nil {
			continue
		}
		for _, location := range asset.Locations {
			if strings.HasPrefix(location, scheme) {
				return location
			}
		}
	}
	return ""
}

// S3 downloads require an AWS instance profile.
func (b *NodeUpScript) useS3Download() bool {
	return b.CloudProvider == string(kops.CloudProviderAWS) && b.firstLocationWithScheme("s3://") != ""
}

// Azure Blob downloads require a managed identity on the instance.
func (b *NodeUpScript) useBlobDownload() bool {
	return b.CloudProvider == string(kops.CloudProviderAzure) && b.firstLocationWithScheme("azureblob://") != ""
}

// ResolveS3Region resolves the bucket region because SigV4 requires it but s3:// URLs omit it.
func (b *NodeUpScript) ResolveS3Region(ctx context.Context, vfsContext *vfs.VFSContext) error {
	if b.CloudProvider != string(kops.CloudProviderAWS) {
		return nil
	}
	location := b.firstLocationWithScheme("s3://")
	if location == "" {
		return nil
	}
	p, err := vfsContext.BuildVfsPath(location)
	if err != nil {
		return fmt.Errorf("building path for %q: %w", location, err)
	}
	s3Path, ok := p.(*vfs.S3Path)
	if !ok {
		return fmt.Errorf("unexpected path type %T for %q", p, location)
	}
	b.S3Region, err = s3Path.Region(ctx)
	if err != nil {
		return fmt.Errorf("getting the region of %q: %w", location, err)
	}
	supported, err := awsup.SupportsS3BootstrapEndpoint(ctx, b.S3Region)
	if err != nil {
		return err
	}
	if !supported {
		return fmt.Errorf("downloading nodeup from an s3:// URL is not supported in AWS region %q", b.S3Region)
	}
	return nil
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

		fmt.Fprintf(writer, "Content-Type: multipart/mixed; boundary=\"%s\"\r\n", boundary)
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

		fmt.Fprintf(writer, "\r\n--%s--\r\n", boundary)

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

		// Map our Insecure Skip Verify setting
		if cluster.Spec.CloudProvider.Openstack != nil && fi.ValueOf(cluster.Spec.CloudProvider.Openstack.InsecureSkipVerify) {
			os.Setenv(openstackconfig.EnvKeyOpenstackTLSInsecureSkipVerify, "true")
		}

		passEnvs := ig.IsControlPlane()
		// Pass in required credentials when using user-defined swift endpoint
		if os.Getenv("OS_AUTH_URL") != "" && passEnvs {
			for _, envVar := range osEnvs {
				env[envVar] = fmt.Sprintf("'%s'", os.Getenv(envVar))
			}
		}
	}

	if cluster.GetCloudProvider() == kops.CloudProviderAWS {
		region, err := awsup.FindRegion(cluster)
		if err != nil {
			return nil, err
		}
		env["AWS_REGION"] = region
	}

	if cluster.GetCloudProvider() == kops.CloudProviderAzure {
		azureEnv := os.Getenv("AZURE_ENVIRONMENT")
		if azureEnv != "" {
			env["AZURE_ENVIRONMENT"] = azureEnv
		}
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
