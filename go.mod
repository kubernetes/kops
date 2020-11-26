module k8s.io/kops

go 1.14

// Version kubernetes-1.20.0-beta.2 => tag v0.20.0-beta.2

replace k8s.io/api => k8s.io/api v0.20.0-beta.2

replace k8s.io/apimachinery => k8s.io/apimachinery v0.20.0-beta.2

replace k8s.io/client-go => k8s.io/client-go v0.20.0-beta.2

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.20.0-beta.2

replace k8s.io/kubectl => k8s.io/kubectl v0.20.0-beta.2

replace k8s.io/apiserver => k8s.io/apiserver v0.20.0-beta.2

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.20.0-beta.2

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.20.0-beta.2

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.20.0-beta.2

replace k8s.io/cri-api => k8s.io/cri-api v0.20.0-beta.2

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.20.0-beta.2

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.20.0-beta.2

replace k8s.io/component-base => k8s.io/component-base v0.20.0-beta.2

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.20.0-beta.2

replace k8s.io/metrics => k8s.io/metrics v0.20.0-beta.2

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.20.0-beta.2

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.20.0-beta.2

replace k8s.io/kubelet => k8s.io/kubelet v0.20.0-beta.2

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.0-beta.2

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.20.0-beta.2

replace k8s.io/code-generator => k8s.io/code-generator v0.20.0-beta.2

replace github.com/gophercloud/gophercloud => github.com/gophercloud/gophercloud v0.11.0

require (
	cloud.google.com/go v0.54.0
	github.com/Azure/azure-pipeline-go v0.2.3
	github.com/Azure/azure-storage-blob-go v0.10.0
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.264
	github.com/aws/amazon-ec2-instance-selector/v2 v2.0.1
	github.com/aws/aws-sdk-go v1.35.24
	github.com/bazelbuild/bazel-gazelle v0.19.1
	github.com/blang/semver/v4 v4.0.0
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/client9/misspell v0.3.4
	github.com/coreos/etcd v3.3.17+incompatible
	github.com/denverdino/aliyungo v0.0.0-20191128015008-acd8035bbb1d
	github.com/digitalocean/godo v1.54.0
	github.com/docker/docker v1.4.2-0.20200309214505-aa6a9891b09c
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20180422025557-ae226422660e
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-ini/ini v1.51.0
	github.com/go-logr/logr v0.2.1-0.20200730175230-ee2de8da5be6
	github.com/gogo/protobuf v1.3.1
	github.com/google/go-cmp v0.5.2
	github.com/google/uuid v1.1.2
	github.com/gophercloud/gophercloud v0.11.1-0.20200518183226-7aec46f32c19
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/hcl/v2 v2.7.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/jacksontj/memberlistmesh v0.0.0-20190905163944-93462b9d2bb7
	github.com/jpillora/backoff v0.0.0-20170918002102-8eab2debe79d
	github.com/kr/fs v0.1.0 // indirect
	github.com/miekg/coredns v0.0.0-20161111164017-20e25559d5ea
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pkg/sftp v0.0.0-20160930220758-4d0e916071f6
	github.com/prometheus/client_golang v1.7.1
	github.com/sergi/go-diff v1.0.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/spotinst/spotinst-sdk-go v1.58.0
	github.com/stretchr/testify v1.6.1
	github.com/urfave/cli v1.22.2
	github.com/weaveworks/mesh v0.0.0-20170419100114-1f158d31de55
	github.com/zclconf/go-cty v1.3.1
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/net v0.0.0-20201110031124-69a78807bb2b
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	golang.org/x/sys v0.0.0-20201112073958-5cba982894dd
	golang.org/x/tools v0.0.0-20200626171337-aa94e735be7f
	google.golang.org/api v0.22.0
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/yaml.v2 v2.3.0
	helm.sh/helm v2.17.0+incompatible
	honnef.co/go/tools v0.0.1-2020.1.4
	k8s.io/api v0.20.0-beta.2
	k8s.io/apimachinery v0.20.0-beta.2
	k8s.io/cli-runtime v0.20.0-beta.2
	k8s.io/client-go v0.20.0-beta.2
	k8s.io/cloud-provider-openstack v1.19.2
	k8s.io/component-base v0.20.0-beta.2
	k8s.io/gengo v0.0.0-20201113003025-83324d819ded
	k8s.io/klog/v2 v2.4.0
	k8s.io/kubectl v0.0.0
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/utils v0.0.0-20201110183641-67b214c5f920
	sigs.k8s.io/controller-runtime v0.6.1
	sigs.k8s.io/controller-tools v0.2.8
	sigs.k8s.io/yaml v1.2.0
)
