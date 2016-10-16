# Copyright 2016 The Kubernetes Authors.
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

#JOB_NAME=kubernetes-e2e-kops-aws
#KUBERNETES_VERSION=v1.3.5
#DNS_DOMAIN="mydomain.com"
#JENKINS_GCS_LOGS_PATH=gs://kopeio-kubernetes-e2e/logs
#KOPS_STATE_STORE=s3://mys3bucket

test: image
	docker run -v ${HOME}/.ssh/id_rsa.pub:/root/.ssh/id_rsa.pub \
	           -v ${HOME}/.aws:/root/.aws \
	           -v ${HOME}/.gsutil:/root/.gsutil \
	           -v ${HOME}/.config/gcloud:/root/.config/gcloud \
	           -e "DNS_DOMAIN=${DNS_DOMAIN}" \
	           -e "JENKINS_GCS_LOGS_PATH=${JENKINS_GCS_LOGS_PATH}" \
	           -e "KOPS_STATE_STORE=${KOPS_STATE_STORE}" \
	           -e "JOB_NAME=${JOB_NAME}" \
	           -e "KUBERNETES_VERSION=${KUBERNETES_VERSION}" \
	           kubernetes-e2e-runner-kops \
	           /runtests.sh

image:
	docker build -t kubernetes-e2e-runner-kops .
