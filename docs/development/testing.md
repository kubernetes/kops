# Testing

## Unit and integration tests

Unit and integration tests can be run using  `go test`

To run all tests:
```
go test -v ./...
```

### Adding an integration test

The integration tests takes a cluster spec and builds cloudformation/terraform templates. To add a new integration test, create a new directory in `tests/integration/update_cluster/` and put the cluster spec in `in-v1alpha2.yaml`. Use a unique cluster name.

Then edit `./cmd/kops/integration_test.go` and add the test function with the cluster name and directory from above.

Lastly run `hack/update-expected` to generate the expected output.

## Kubernetes e2e testing

### Preparing the environment

The easiest way to run the Kubernetes e2e tests is to install the `kubetest` binary from test-infra.

See [e2e-tests](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-testing/e2e-tests.md) for instructions how to install and general usage of the utility.

You can see https://github.com/kubernetes/test-infra/blob/master/config/jobs/kubernetes/kops/ the various jobs that are periodically run against kops.
The container arguments for each of the jobs are the ones passed to `kubetest`.

Following the examples below, kubetest will download artifacts such as a given Kubernetes build. Therefore you probably want to run `kubetest` from a directory dedicated for this purpose.

### Running against an existing cluster

If you want `kubetest` to run spin up and tear down the cluster on every run, you can run something like the following to re-use an existing cluster.
This assumes you have already built the kops binary from source.

```
GINKGO_PARALLEL=y kubetest --test --test_args="--ginkgo.skip=\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|Services.*functioning.*NodePort" --provider=aws --deployment=kops --cluster=my.testcluster.com --kops-state=${KOPS_STATE_STORE} --kops ${GOPATH}/bin/kops --extract=ci/latest
```

Note the `--extract` flag is only needed on first run or whenever you want to update the kubernetes build. You can omit this flag to run against the already extacted tree on subsequent runs. Just `cd` into the `kubernetes` directory first.

### Running against a new cluster

By adding the `--up` flag, `kubetest` will also spin up a new cluster. In most cases, you also need to add a few additional flags.

```
GINKGO_PARALLEL=y kubetest --test --test_args="--ginkgo.skip=\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|Services.*functioning.*NodePort" --provider=aws --check-version-skew=false --deployment=kops --kops-state=${KOPS_STATE_STORE} --kops ${GOPATH}/bin/kops --kops-args="--network-cidr=192.168.1.0/24" --cluster=my.testcluster.com --up --kops-ssh-key ~/.ssh/id_rsa --kops-admin-access=0.0.0.0/0
```

If you want to run the tests against your development version of kops, you need to upload the binaries and set the environment variables as described in [Adding a new feature](adding_a_feature.md).

Since we assume you are using this cluster for testing, we leave the cluster running after the tests have finished so that you can inspect the nodes if anything unexpected happens. If you do not need this, you can add the `--down` flag. Otherwise, just delete the cluster as any other cluster: `kops delete cluster my.testcluster.com --yes`
