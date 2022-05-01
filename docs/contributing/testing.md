# Testing

## Unit and integration tests

Unit and integration tests can be run using  `go test`

To run all tests:
```
go test -v ./...
```

### Adding an integration test

The integration tests takes a cluster spec and builds cloudformation/terraform templates. For new functionality, consider adding it to the `complex.example.com` cluster unless it conflicts with existing functionality in that cluster. To add a new integration test, create a new directory in `tests/integration/update_cluster/` and put the cluster spec in `in-v1alpha2.yaml`. Use a unique cluster name.

Then edit `./cmd/kops/integration_test.go` and add the test function with the cluster name and directory from above.

Lastly run `./hack/update-expected.sh` to generate the expected output.

## Kubernetes e2e testing

Kubetest2 is the framework for launching and running end-to-end tests on Kubernetes, and the best approach to test your kOps cluster is to use the same Go modules to perform the e2e testing.

### Preparing the environment

Before running `kubetest2` you will need to install the core, and all deployers and testers Go modules. 

```shell
make test-e2e-install
```

For reference, the build target commands can be found [here](https://github.com/kubernetes/kops/tree/master/tests/e2e/e2e.mk).

See [GitHub kubetest2](https://github.com/kubernetes-sigs/kubetest2/blob/master/README.md) to gain a further understanding of the Kubernetes e2e test framework.

Following the examples below, `kubetest2` will download test artifacts to `./_artifacts`.

### Running against an existing cluster

You can run something like the following to have `kubetest2` re-use an existing cluster.

This assumes you have already built the kOps binary from source. The exact path to the `kops` binary used in the `--kops-binary-path` flag may differ.

The environment variable `KOPS_ROOT` is the full path to your local GitHub kOps working directory.   

```shell
kubetest2 kops \
  -v 2 \
  --test \
  --cloud-provider=aws \
  --cluster-name=my.testcluster.com \
  --kops-binary-path=${KOPS_ROOT}/.build/dist/$(go env GOOS)/$(go env GOARCH)/kops \
  --kubernetes-version=v1.20.2 \
  --test=kops \
  -- \
  --test-package-version=v1.20.2 \
  --parallel 25 \
  --skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler"
```

It's also possible to run the Kubernetes Conformance test suite by replacing the `--skip-regex` flag with `--focus-regex='\[Conformance\]'`.

See [Conformance Testing in Kubernetes](https://github.com/kubernetes/community/blob/master/contributors/devel/sig-architecture/conformance-tests.md)

### Running against a new cluster

By adding the `--up` flag, `kubetest2` will spin up a new cluster. Flags can be passed to the `kops create cluster` command via the `--create-args` flag. In most cases, you also need to add a few additional flags. See `kubetest2 kops --help` for the full list.

```shell
kubetest2 kops \
  -v 2 \
  --up \
  --cloud-provider=aws \
  --cluster-name=my.testcluster.com \
  --create-args="--networking calico" \
  --kops-binary-path=${KOPS_ROOT}/.build/dist/$(go env GOOS)/$(go env GOARCH)/kops \
  --kubernetes-version=v1.20.2 \
  --test=kops \
  --
  -- \
  --test-package-version=v1.20.2 \
  --parallel 25 \
  --skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler"
```

If you don't specify any additional flags, the kOps deployer Go module will create a kubernetes cluster using the following defaults.

```shell
kops create cluster --name my.testcluster.com --admin-access <Client Public IP> --cloud aws --kubernetes-version v1.20.2 --master-count 1 --master-volume-size 48 --node-count 4 --node-volume-size 48 --override cluster.spec.nodePortAccess=0.0.0.0/0 --ssh-public-key /home/ubuntu/.ssh/id_rsa.pub --yes --zones <Random Zone> --master-size c5.large --networking calico
```

For the `--zones` flag, the kOps deployer will select a random zone based on the `--cloud-provider` flag, for `aws` the full list of AWS zones can be found [here](https://github.com/kubernetes/kops/blob/master/tests/e2e/kubetest2-kops/aws/zones.go) and for `gce` the full list of GCE zones can be found [here](https://github.com/kubernetes/kops/blob/master/tests/e2e/kubetest2-kops/gce/zones.go).

Althernatively, you can generate a kOps cluster spec YAML manifest based on your own requirments using `kops create cluster my.testcluster.com ... --dry-run -oyaml > my.testcluster.com.yaml` and then run the `kubetest2` e2e tests using the `--template-path` flag to specify the full path to the YAML manifest.

```shell
kubetest2 kops \
  -v 2 \
  --up \
  --cloud-provider=aws \
  --cluster-name=my.testcluster.com \
  --kops-binary-path=${KOPS_ROOT}/.build/dist/$(go env GOOS)/$(go env GOARCH)/kops \
  --kubernetes-version=v1.20.2 \
  --template-path=my.testcluster.com.yaml \
  --test=kops \
  -- \
  --test-package-version=v1.20.2 \
  --parallel 25 \
  --skip-regex="\[Slow\]|\[Serial\]|\[Disruptive\]|\[Flaky\]|\[Feature:.+\]|\[HPA\]|Dashboard|RuntimeClass|RuntimeHandler"
```

If you encounter the following error, you will need to add your SSH public key to the kOps cluster spec YAML manifest.

```shell
SSH public key must be specified when running with AWS (create with `kops create secret --name training.kops.k8s.local sshpublickey admin -i ~/.ssh/id_rsa.pub`)
```

```yaml
---

apiVersion: kops.k8s.io/v1alpha2
kind: SSHCredential
metadata:
  name: admin
  labels:
    kops.k8s.io/cluster: my.testcluster.com
spec:
  publicKey: "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAACAQC..."
```

If you want to run the tests against your development version of kOps, you need to upload the binaries and set the environment variables as described in [Adding a new feature](adding_a_feature.md#testing).

Since we assume you are using this cluster for testing, we leave the cluster running after the tests have finished so that you can inspect the nodes if anything unexpected happens. If you do not need this, you can add the `--down` flag. Otherwise, just delete the cluster as any other cluster: `kops delete cluster my.testcluster.com --yes`
