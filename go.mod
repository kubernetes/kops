module k8s.io/kops

go 1.16

// Version kubernetes-1.21.0-beta.1 => tag v0.21.0-beta.1

// This should match hack/go.mod
replace k8s.io/code-generator => k8s.io/code-generator v0.21.0-beta.1

replace (
	k8s.io/api => k8s.io/api v0.21.0-beta.1
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.0-beta.1
	k8s.io/apimachinery => k8s.io/apimachinery v0.21.0-beta.1
	k8s.io/apiserver => k8s.io/apiserver v0.21.0-beta.1
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.0-beta.1
	k8s.io/client-go => k8s.io/client-go v0.21.0-beta.1
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.0-beta.1
	k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.0-beta.1
	k8s.io/component-base => k8s.io/component-base v0.21.0-beta.1
	k8s.io/component-helpers => k8s.io/component-helpers v0.21.0-beta.1
	k8s.io/controller-manager => k8s.io/controller-manager v0.21.0-beta.1
	k8s.io/cri-api => k8s.io/cri-api v0.21.0-beta.1
	k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.0-beta.1
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.0-beta.1
	k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.0-beta.1
	k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.0-beta.1
	k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.0-beta.1
	k8s.io/kubectl => k8s.io/kubectl v0.21.0-beta.1
	k8s.io/kubelet => k8s.io/kubelet v0.21.0-beta.1
	k8s.io/kubernetes => k8s.io/kubernetes v1.21.0-beta.1
	k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.0-beta.1
	k8s.io/metrics => k8s.io/metrics v0.21.0-beta.1
	k8s.io/mount-utils => k8s.io/mount-utils v0.21.0-beta.1
	k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.0-beta.1
)

require (
	cloud.google.com/go v0.77.0
	github.com/Azure/azure-pipeline-go v0.2.3
	github.com/Azure/azure-sdk-for-go v48.2.0+incompatible
	github.com/Azure/azure-storage-blob-go v0.10.0
	github.com/Azure/go-autorest/autorest v0.11.12
	github.com/Azure/go-autorest/autorest/azure/auth v0.5.3
	github.com/Azure/go-autorest/autorest/to v0.4.0
	github.com/MakeNowJust/heredoc/v2 v2.0.1
	github.com/Masterminds/sprig/v3 v3.1.0
	github.com/aliyun/alibaba-cloud-sdk-go v1.61.264
	github.com/aws/amazon-ec2-instance-selector/v2 v2.0.1
	github.com/aws/aws-sdk-go v1.37.11
	github.com/blang/semver/v4 v4.0.0
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/denverdino/aliyungo v0.0.0-20191128015008-acd8035bbb1d
	github.com/digitalocean/godo v1.54.0
	github.com/docker/docker v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/go-bindata/go-bindata/v3 v3.1.3
	github.com/go-ini/ini v1.62.0
	github.com/go-logr/logr v0.4.0
	github.com/gogo/protobuf v1.3.2
	github.com/google/go-cmp v0.5.4
	github.com/google/uuid v1.2.0
	github.com/googleapis/gnostic v0.5.4 // indirect
	github.com/gophercloud/gophercloud v0.15.0
	github.com/hashicorp/hcl/v2 v2.7.0
	github.com/hashicorp/vault/api v1.0.4
	github.com/imdario/mergo v0.3.11 // indirect
	github.com/jacksontj/memberlistmesh v0.0.0-20190905163944-93462b9d2bb7
	github.com/jetstack/cert-manager v1.1.0
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pelletier/go-toml v1.8.1
	github.com/pkg/sftp v1.12.0
	github.com/prometheus/client_golang v1.9.0
	github.com/prometheus/common v0.17.0 // indirect
	github.com/prometheus/procfs v0.6.0 // indirect
	github.com/sergi/go-diff v1.1.0
	github.com/spf13/cobra v1.1.1
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0
	github.com/spotinst/spotinst-sdk-go v1.80.0
	github.com/stretchr/testify v1.7.0
	github.com/weaveworks/mesh v0.0.0-20170419100114-1f158d31de55
	github.com/zclconf/go-cty v1.3.1
	golang.org/x/crypto v0.0.0-20210220033148-5ea612d1eb83
	golang.org/x/net v0.0.0-20210224082022-3d97a244fca7
	golang.org/x/oauth2 v0.0.0-20210220000619-9bb904979d93
	golang.org/x/sys v0.0.0-20210225134936-a50acf3fe073
	google.golang.org/api v0.40.0
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	helm.sh/helm/v3 v3.4.2
	k8s.io/api v0.21.0-beta.1
	k8s.io/apiextensions-apiserver v0.21.0-beta.1 // indirect
	k8s.io/apimachinery v0.21.0-beta.1
	k8s.io/cli-runtime v0.21.0-beta.1
	k8s.io/client-go v0.21.0-beta.1
	k8s.io/cloud-provider-openstack v1.20.1
	k8s.io/component-base v0.21.0-beta.1
	k8s.io/gengo v0.0.0-20201214224949-b6c5ce23f027
	k8s.io/klog/v2 v2.5.0
	k8s.io/kubectl v0.21.0-beta.1
	k8s.io/legacy-cloud-providers v0.21.0-beta.1
	k8s.io/utils v0.0.0-20210111153108-fddb29f9d009
	sigs.k8s.io/controller-runtime v0.8.2
	sigs.k8s.io/yaml v1.2.0
)
