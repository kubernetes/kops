## Run kubernetes e2e tests

This docker image lets you run the kubernetes e2e tests very easily, using kops to create the cluster.

You simply call make, specifying some variables that controls the build.

An example:

`make JOB_NAME=kubernetes-e2e-kops-aws KUBERNETES_VERSION=v1.3.5 DNS_DOMAIN=e2e.mydomain.com JENKINS_GCS_LOGS_PATH=gs://kopeio-kubernetes-e2e/logs KOPS_STATE_STORE=s3://clusters.mydomain.com`

Variables:

* `JOB_NAME` the e2e job to run.  Corresponds to a conf file in the conf directory.
* `KUBERNETES_VERSION` the version of kubernetes to run.  Either a version like `v1.3.5`, or a URL prefix like `https://storage.googleapis.com/kubernetes-release-dev/ci/v1.4.0-alpha.2.677+ea69570f61af8e/`.  See [testing docs](../docs/testing.md)
* `DNS_DOMAIN` the dns domain name to use for the cluster.  Must be a real domain name, with a zone registered in DNS (route53)
* `JENKINS_GCS_LOGS_PATH` the gs bucket where we should upload the results of the build.  Note these will be publicly readable.
* `KOPS_STATE_STORE` the url where the kops registry (store of cluster information) lives.

