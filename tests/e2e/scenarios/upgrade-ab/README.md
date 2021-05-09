### Running locally

If you want run locally:

First set your KOPS_STATE_STORE:

```
export KOPS_STATE_STORE=...
```

Then set the configuration for the test run:

```
export KOPS_VERSION_A="v1.18.3"
export KOPS_VERSION_B="v1.19.2"
export K8S_VERSION_A="v1.18.18"
export K8S_VERSION_B="v1.18.18"
export ADMIN_ACCESS="0.0.0.0/0" # Or use your IPv4 with /32

export CLOUD_PROVIDER=aws
export CLUSTER_NAME=upgrade-ab.k8s.local

export PATH=${GOPATH}/bin:$PATH

tests/e2e/scenarios/upgrade-ab/run-test.sh
```
