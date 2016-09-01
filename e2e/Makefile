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

