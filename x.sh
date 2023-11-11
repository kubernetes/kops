#!/bin/bash
set -o errexit
set -o nounset
set -o pipefail

NODEUP_URL_AMD64=https://storage.googleapis.com/k8s-staging-kops/pulls/pull-kops-e2e-cni-calico/pull-f7a52388b66079ae2487a2a1af4fe8d55549e1ee/1.29.0-alpha.2+v1.29.0-alpha.1-167-g6b543cd2f7/linux/amd64/nodeup
NODEUP_HASH_AMD64=f46b20a651727d81b1bdf625648f53a9564325610b1446cb292c0253f34e84b5
NODEUP_URL_ARM64=
NODEUP_HASH_ARM64=

export AWS_REGION=us-west-1




sysctl -w net.core.rmem_max=16777216 || true
sysctl -w net.core.wmem_max=16777216 || true
sysctl -w net.ipv4.tcp_rmem='4096 87380 16777216' || true
sysctl -w net.ipv4.tcp_wmem='4096 87380 16777216' || true


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
  echo "Downloading \"$1\" from \"$3\" with hash \"$2\""
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
        "curl -f --compressed -Lo \"${file}\" --connect-timeout 20 --retry 6 --retry-delay 10"
        "wget --compression=auto -O \"${file}\" --connect-timeout=20 --tries=6 --wait=10"
        "curl -f -Lo \"${file}\" --connect-timeout 20 --retry 6 --retry-delay 10"
        "wget -O \"${file}\" --connect-timeout=20 --tries=6 --wait=10"
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

  actual=$(sha256sum "${file}" | awk '{ print $1 }') || true
  if [[ "${actual}" != "${expected}" ]]; then
    echo "== ${file} corrupted, hash ${actual} doesn't match expected ${expected} =="
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

  echo "Running nodeup"
  # We can't run in the foreground because of https://github.com/docker/docker/issues/23793
  ( cd ${INSTALL_DIR}/bin; ./nodeup --install-systemd-unit --conf=${INSTALL_DIR}/conf/kube_env.yaml --v=8  )
}

####################################################################################

/bin/systemd-machine-id-setup || echo "failed to set up ensure machine-id configured"

echo "== nodeup node config starting =="
ensure-install-dir

cat > conf/kube_env.yaml << '__EOF_KUBE_ENV'
CloudProvider: aws
ClusterName: e2e-pr16044.pull-kops-e2e-cni-calico.test-cncf-aws.k8s.io
ConfigServer:
  CACertificates: |
    -----BEGIN CERTIFICATE-----
    MIIC+DCCAeCgAwIBAgIMF5Cqac57lJxbLnE0MA0GCSqGSIb3DQEBCwUAMBgxFjAU
    BgNVBAMTDWt1YmVybmV0ZXMtY2EwHhcNMjMxMDIxMDcwMDMxWhcNMzMxMDIwMDcw
    MDMxWjAYMRYwFAYDVQQDEw1rdWJlcm5ldGVzLWNhMIIBIjANBgkqhkiG9w0BAQEF
    AAOCAQ8AMIIBCgKCAQEAp0c0uZDx27thN/awFEFylpBPc0mgclX0G04mmQbX+4K+
    cZROlQRo2PAijFK9HhccboLumvGJOlUFyPS4/F8PdPVgJctSVg9abz23RjGeRQal
    fopoEE6x33/Mfyee+n1P9aHP7pEre8Sb4mKGraYOfzIHwDhl9zBoweAERYuEVEur
    Ig0g5/NL+81jaVAm3kPeDOFkg6aUIpMS2MJJRbGr4Bm7V/xVxrUof454JcgcXSM+
    11FjcZDL/qRNQ8r3nDepc57eBWOo/Wi7fPqiztF5Qk6a5s2b9ZM1u+LcX854jIzQ
    WkBNzxKdNATG+lhYmwWSIKp7xYMhxJE69y5+xkF5mQIDAQABo0IwQDAOBgNVHQ8B
    Af8EBAMCAQYwDwYDVR0TAQH/BAUwAwEB/zAdBgNVHQ4EFgQUEUdui0mKJvkTg35s
    cS4/Ow85yAUwDQYJKoZIhvcNAQELBQADggEBAISN06W0+ivx03G366mL1QqGpKpR
    xzCcyT0C7AnNVHdQoYqxw0HjFgIVHW/QpDbXzUfMQT34vl8gSVkLdbBr8jbizFBH
    4Y9alxHvrgy4y/kazKMNXui43BTUEfWlULRWO7xKLv8UvV5mgxhxcbslGwhO0VVj
    oNTZAdXpV6GkwdDQ18dXLOvzV++ojdsRu/dWvJY9Ft7gBO/gHUygPP8P5vHZLG7l
    JJpVOyjp7PZYd5pXd0UEegIH9mMl2xHShW06KTHaDR4l/3N5i2KrQGjC9fVfM5Ce
    pbdNf6ENEry8MxkwzntySM69MlDMeA2Whsx4LfXjBVIAuv3DhoPIMCKeczc=
    -----END CERTIFICATE-----
  servers:
  - https://kops-controller.internal.e2e-pr16044.pull-kops-e2e-cni-calico.test-cncf-aws.k8s.io:3988/
InstanceGroupName: nodes-us-west-1a
InstanceGroupRole: Node
NodeupConfigHash: cGC5Y/AAHLZhAYkwIIp5EXzfhfUOet7ci1kRPsFjTkA=

__EOF_KUBE_ENV

download-release
echo "== nodeup node config done =="