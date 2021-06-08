module k8s.io/kops/tests/e2e

go 1.16

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/octago/sflags v0.2.0
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/apimachinery v0.21.1
	k8s.io/klog/v2 v2.8.0
	k8s.io/kops v0.0.0-00010101000000-000000000000
	sigs.k8s.io/boskos v0.0.0-20200710214748-f5935686c7fc
	sigs.k8s.io/kubetest2 v0.0.0-20210423234514-1c731a5d2283
)

replace k8s.io/kops => ../../.

// These should match the go.mod from k8s.io/kops

replace k8s.io/api => k8s.io/api v0.21.0

replace k8s.io/apimachinery => k8s.io/apimachinery v0.21.0

replace k8s.io/client-go => k8s.io/client-go v0.21.0

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.0

replace k8s.io/controller-manager => k8s.io/controller-manager v0.21.0

replace k8s.io/kubectl => k8s.io/kubectl v0.21.0

replace k8s.io/apiserver => k8s.io/apiserver v0.21.0

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.0

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.0

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.0

replace k8s.io/cri-api => k8s.io/cri-api v0.21.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.0

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.0

replace k8s.io/component-base => k8s.io/component-base v0.21.0

replace k8s.io/component-helpers => k8s.io/component-helpers v0.21.0

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.0

replace k8s.io/metrics => k8s.io/metrics v0.21.0

replace k8s.io/mount-utils => k8s.io/mount-utils v0.21.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.0

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.0

replace k8s.io/kubelet => k8s.io/kubelet v0.21.0

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.0

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.0

replace k8s.io/code-generator => k8s.io/code-generator v0.21.0
