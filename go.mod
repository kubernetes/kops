module k8s.io/kops

go 1.16

// Version kubernetes-1.21.1 => tag v0.21.1

// This should match hack/go.mod
replace k8s.io/code-generator => k8s.io/code-generator v0.21.1

replace (
	k8s.io/api => k8s.io/api v0.21.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.1
	k8s.io/apiserver => k8s.io/apiserver v0.21.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.1
	k8s.io/client-go => k8s.io/client-go v0.21.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.1
	k8s.io/component-base => k8s.io/component-base v0.21.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.1
	k8s.io/cri-api => k8s.io/cri-api v0.21.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.1
	k8s.io/kubectl => k8s.io/kubectl v0.21.1
	k8s.io/kubelet => k8s.io/kubelet v0.21.1
	k8s.io/kubernetes => k8s.io/kubernetes v1.21.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.1
	k8s.io/metrics => k8s.io/metrics v0.21.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.1
)

require (
	cloud.google.com/go v0.81.0
	github.com/Azure/azure-pipeline-go v0.2.3
	github.com/Azure/azure-sdk-for-go v53.4.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.13.0
	github.com/Azure/go-autorest/autorest v0.11.18
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.7
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/Masterminds/sprig/v3 v3.2.2
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.1059
	github.com/apparentlymart/go-cidr v1.1.0
	github.com/aws/amazon-ec2-instance-selector/v2 v2.0.2
	github.com/aws/aws-sdk-go v1.40.0
	github.com/blang/semver/v4 v4.0.0
	github.com/denverdino/aliyungo v0.0.0-20210425065611-55bee4942cba
	github.com/digitalocean/godo v1.60.0
	github.com/docker/docker v20.10.6+incompatible // indirect
	github.com/go-ini/ini v1.62.0
	github.com/go-logr/logr v0.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/google/go-cmp v0.5.5
	github.com/google/go-containerregistry v0.5.1
	github.com/google/uuid v1.2.0
	github.com/gophercloud/gophercloud v0.18.0
	github.com/hashicorp/hcl/v2 v2.10.0
	github.com/hashicorp/vault/api v1.1.0
	github.com/jacksontj/memberlistmesh v0.0.0-20190905163944-93462b9d2bb7
	github.com/jetstack/cert-manager v1.3.1
	github.com/mitchellh/mapstructure v1.4.1
	github.com/pelletier/go-toml v1.9.3
	github.com/pkg/sftp v1.13.0
	github.com/prometheus/client_golang v1.11.0
	github.com/sergi/go-diff v1.2.0
	github.com/spf13/cobra v1.2.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.8.1
	github.com/spotinst/spotinst-sdk-go v1.85.0
	github.com/stretchr/testify v1.7.0
	github.com/weaveworks/mesh v0.0.0-20191105120815-58dbcc3e8e63
	github.com/zclconf/go-cty v1.8.2
	golang.org/x/crypto v0.0.0-20210513164829-c07d793c2f9a
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e
	golang.org/x/oauth2 v0.0.0-20210514164344-f6687ab2804c
	golang.org/x/sys v0.0.0-20210603081109-ebe580a85c40
	google.golang.org/api v0.45.0
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/square/go-jose.v2 v2.5.1
	helm.sh/helm/v3 v3.6.1
	k8s.io/api v0.21.1
	k8s.io/apimachinery v0.21.1
	k8s.io/cli-runtime v0.21.1
	k8s.io/client-go v0.21.1
	k8s.io/cloud-provider-openstack v1.21.1
	k8s.io/component-base v0.21.1
	k8s.io/gengo v0.0.0-20210203185629-de9496dff47b
	k8s.io/klog/v2 v2.8.0
	k8s.io/kubectl v0.21.1
	k8s.io/legacy-cloud-providers v0.21.1
	k8s.io/mount-utils v0.21.1
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/controller-runtime v0.9.0
	sigs.k8s.io/yaml v1.2.0
)
