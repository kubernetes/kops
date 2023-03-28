### Running locally

If you want run locally:

First set your KOPS_STATE_STORE:

```
export KOPS_STATE_STORE=...
```

Then set the configuration and start the test run:

```
export CLUSTER_NAME=scalability.k8s.local
# ... and any other values you want to override

tests/e2e/scenarios/scalability/run-test.sh
```
