#!/bin/bash -ex

if [[ -z "${JOB_NAME}" ]]; then
  echo "Must specify JOB_NAME env var"
  exit 1
fi
if [[ -z "${KUBERNETES_VERSION}" ]]; then
  echo "Must specify KUBERNETES_VERSION env var"
  exit 1
fi
if [[ -z "${DNS_DOMAIN}" ]]; then
  echo "Must specify DNS_DOMAIN env var"
  exit 1
fi
if [[ -z "${KOPS_STATE_STORE}" ]]; then
  echo "Must specify KOPS_STATE_STORE env var"
  exit 1
fi
# TODO: Maybe skip if we don't want to upload logs?
if [[ -z "${JENKINS_GCS_LOGS_PATH}" ]]; then
  echo "Must specify JENKINS_GCS_LOGS_PATH env var"
  exit 1
fi

echo "JOB_NAME=${JOB_NAME}"
echo "Loading conf/${JOB_NAME}"

. conf/${JOB_NAME}

echo "Loading conf/cloud/${KUBERNETES_PROVIDER}"
. conf/cloud/${KUBERNETES_PROVIDER}

echo "Loading conf/site"
. conf/site

##=============================================================
# Global settings
export KUBE_GCS_RELEASE_BUCKET=kubernetes-release

# We download the binaries ourselves
# TODO: No way to tell e2e to use a particular release?
# TODO: Maybe download and then bring up the cluster?
export JENKINS_USE_EXISTING_BINARIES=y

# This actually just skips kube-up master detection
export KUBERNETES_CONFORMANCE_TEST=y

##=============================================================
# System settings (emulate jenkins)
export USER=root
export WORKSPACE=$HOME
# Nothing should want Jenkins $HOME
export HOME=${WORKSPACE}
export BUILD_NUMBER=`date -u +%Y%m%d%H%M%S`
export JENKINS_HOME=${HOME}

# We'll directly up & down the cluster
export E2E_UP="${E2E_UP:-false}"
export E2E_TEST="${E2E_TEST:-true}"
export E2E_DOWN="${E2E_DOWN:-false}"

# Skip gcloud update checking
export CLOUDSDK_COMPONENT_MANAGER_DISABLE_UPDATE_CHECK=true


##=============================================================

branch=master

build_dir=${JENKINS_HOME}/jobs/${JOB_NAME}/builds/${BUILD_NUMBER}/
rm -rf ${build_dir}
mkdir -p ${build_dir}/workspace

cd ${build_dir}/workspace

# Sanity check
#gsutil ls ${JENKINS_GCS_LOGS_PATH}

exit_code=0
SECONDS=0 # magic bash timer variable
curl -fsS --retry 3  "https://raw.githubusercontent.com/kubernetes/kubernetes/master/hack/jenkins/e2e-runner.sh" > /tmp/e2e.sh
chmod +x /tmp/e2e.sh

# We need kubectl to write kubecfg from kops
curl -fsS --retry 3  "https://storage.googleapis.com/kubernetes-release/release/v1.3.5/bin/linux/amd64/kubectl" > /usr/local/bin/kubectl
chmod +x /usr/local/bin/kubectl

curl -fsS --retry 3  "https://kubeupv2.s3.amazonaws.com/kops/kops-1.3.tar.gz" > /tmp/kops.tar.gz
tar zxf /tmp/kops.tar.gz -C /opt

if [[ ! -e ${AWS_SSH_KEY} ]]; then
  echo "Creating ssh key ${AWS_SSH_KEY}"
  ssh-keygen -N "" -t rsa -f ${AWS_SSH_KEY}
fi

function fetch_tars_from_gcs() {
    local -r bucket="${1}"
    local -r build_version="${2}"
    echo "Pulling binaries from GCS; using server version ${bucket}/${build_version}."
    gsutil -mq cp \
        "gs://${KUBE_GCS_RELEASE_BUCKET}/${bucket}/${build_version}/kubernetes.tar.gz" \
        "gs://${KUBE_GCS_RELEASE_BUCKET}/${bucket}/${build_version}/kubernetes-test.tar.gz" \
        .
}

function unpack_binaries() {
    md5sum kubernetes*.tar.gz
    tar -xzf kubernetes.tar.gz
    tar -xzf kubernetes-test.tar.gz
}


fetch_tars_from_gcs release ${KUBERNETES_VERSION}
unpack_binaries

# Clean up everything when we're done
function finish {
  /opt/kops/kops delete cluster \
                 --name ${JOB_NAME}.${DNS_DOMAIN} \
                 --yes  2>&1 | tee -a ${build_dir}/build-log.txt
}
trap finish EXIT

set -e

# Create the cluster spec
pushd /opt/kops
/opt/kops/kops create cluster \
                  --name ${JOB_NAME}.${DNS_DOMAIN} \
                  --cloud ${KUBERNETES_PROVIDER} \
                  --zones ${NODE_ZONES} \
                  --node-size ${NODE_SIZE} \
                  --master-size ${MASTER_SIZE} \
                  --ssh-public-key ${AWS_SSH_KEY}.pub \
                  --kubernetes-version ${KUBERNETES_VERSION} \
                  --v=4 2>&1 | tee -a ${build_dir}/build-log.txt
exit_code=${PIPESTATUS[0]}
popd

# Apply the cluster spec
if [[ ${exit_code} == 0 ]]; then
  pushd /opt/kops
  /opt/kops/kops update cluster \
                    --name ${JOB_NAME}.${DNS_DOMAIN} \
                    --yes --v=4 2>&1 | tee -a ${build_dir}/build-log.txt
  exit_code=${PIPESTATUS[0]}
  popd
fi

# Wait for kubectl to begin responding (at least master up)
if [[ ${exit_code} == 0 ]]; then
  attempt=0
  while true; do
    kubectl get nodes --show-labels  2>&1 | tee -a ${build_dir}/build-log.txt
    exit_code=${PIPESTATUS[0]}

    if [[ ${exit_code} == 0 ]]; then
      break
    fi
    if (( attempt > 60 )); then
      echo "Unable to connect to API in 15 minutes (master did not launch?)"
      break
    fi
    attempt=$(($attempt+1))
    sleep 15
  done
fi

# TODO: can we get rid of this?
echo "API responded; waiting 450 seconds for DNS to settle"
for ((i=1;i<=15;i++)); do
    kubectl get nodes --show-labels  2>&1 | tee -a ${build_dir}/build-log.txt
    sleep 30
done


# Run e2e tests
if [[ ${exit_code} == 0 ]]; then
  /tmp/e2e.sh 2>&1 | tee -a ${build_dir}/build-log.txt
  exit_code=${PIPESTATUS[0]}
fi

# Try to clean up normally so it goes into the logs
# (we have an exit hook for abnormal termination, but that does not get logged)
finish

duration=$SECONDS
set +e

if [[ ${exit_code} == 0 ]]; then
  success="true"
else
  success="false"
fi

version=`cat kubernetes/version`

gcs_acl="public-read"
gcs_job_path="${JENKINS_GCS_LOGS_PATH}/${JOB_NAME}"
gcs_build_path="${gcs_job_path}/${BUILD_NUMBER}"

gsutil -q cp -a "${gcs_acl}" -z txt "${build_dir}/build-log.txt" "${gcs_build_path}/"

curl -fsS --retry 3 "https://raw.githubusercontent.com/kubernetes/kubernetes/master/hack/jenkins/upload-to-gcs.sh" | bash -


curl -fsS --retry 3 "https://raw.githubusercontent.com/kubernetes/kubernetes/master/hack/jenkins/upload-finished.sh" > upload-finished.sh
chmod +x upload-finished.sh

if [[ ${exit_code} == 0 ]]; then
  ./upload-finished.sh SUCCESS
else
  ./upload-finished.sh UNSTABLE
fi

