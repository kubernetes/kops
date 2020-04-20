module k8s.io/kops

go 1.13

// Version kubernetes-1.18.0 => tag v0.18.1

replace k8s.io/api => k8s.io/api v0.18.1

replace k8s.io/apimachinery => k8s.io/apimachinery v0.18.1

replace k8s.io/client-go => k8s.io/client-go v0.18.1

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.18.1

replace k8s.io/kubectl => k8s.io/kubectl v0.18.1

replace k8s.io/apiserver => k8s.io/apiserver v0.18.1

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.18.1

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.18.1

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.18.1

replace k8s.io/cri-api => k8s.io/cri-api v0.18.1

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.18.1

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.18.1

replace k8s.io/component-base => k8s.io/component-base v0.18.1

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.18.1

replace k8s.io/metrics => k8s.io/metrics v0.18.1

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.18.1

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.18.1

replace k8s.io/kubelet => k8s.io/kubelet v0.18.1

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.1

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.18.1

replace k8s.io/code-generator => k8s.io/code-generator v0.18.1

replace github.com/gophercloud/gophercloud => github.com/gophercloud/gophercloud v0.9.0

require (
	cloud.google.com/go v0.38.0
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/semver v1.3.1 // indirect
	github.com/Masterminds/sprig v2.17.1+incompatible
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/aws/aws-sdk-go v1.29.21
	github.com/bazelbuild/bazel-gazelle v0.19.1
	github.com/blang/semver v3.5.0+incompatible
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/client9/misspell v0.3.4
	github.com/coreos/etcd v3.3.17+incompatible
	github.com/denverdino/aliyungo v0.0.0-20191128015008-acd8035bbb1d
	github.com/digitalocean/godo v1.19.0
	github.com/docker/engine-api v0.0.0-20160509170047-dea108d3aa0c
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20180422025557-ae226422660e
	github.com/ghodss/yaml v1.0.0
	github.com/go-bindata/go-bindata v3.1.2+incompatible
	github.com/go-ini/ini v1.51.0
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.3
	github.com/gophercloud/gophercloud v0.7.1-0.20200116011225-46fdd1830e9a
	github.com/gorilla/mux v1.7.3
	github.com/hashicorp/hcl v1.0.0
	github.com/hashicorp/hcl/v2 v2.3.0
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/jacksontj/memberlistmesh v0.0.0-20190905163944-93462b9d2bb7
	github.com/jpillora/backoff v0.0.0-20170918002102-8eab2debe79d
	github.com/kr/fs v0.1.0 // indirect
	github.com/miekg/coredns v0.0.0-20161111164017-20e25559d5ea
	github.com/miekg/dns v1.1.16
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v0.0.0-20160930220758-4d0e916071f6
	github.com/prometheus/client_golang v1.0.0
	github.com/sergi/go-diff v1.0.0
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.4.0
	github.com/spotinst/spotinst-sdk-go v1.43.0
	github.com/stretchr/testify v1.4.0
	github.com/urfave/cli v1.20.0
	github.com/vmware/govmomi v0.20.3
	github.com/weaveworks/mesh v0.0.0-20170419100114-1f158d31de55
	github.com/zclconf/go-cty v1.3.1
	go.uber.org/zap v1.10.0
	golang.org/x/crypto v0.0.0-20200220183623-bac4c82f6975
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sys v0.0.0-20191128015809-6d18c012aee9
	golang.org/x/tools v0.0.0-20191203134012-c197fd4bf371
	google.golang.org/api v0.17.0
	gopkg.in/gcfg.v1 v1.2.3
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/yaml.v2 v2.2.8
	honnef.co/go/tools v0.0.1-2019.2.3
	k8s.io/api v0.18.1
	k8s.io/apimachinery v0.18.1
	k8s.io/cli-runtime v0.18.1
	k8s.io/client-go v0.18.1
	k8s.io/cloud-provider-openstack v1.17.0
	k8s.io/component-base v0.18.1
	k8s.io/helm v2.9.0+incompatible
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.0.0
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/utils v0.0.0-20200324210504-a9aa75ae1b89
	sigs.k8s.io/controller-runtime v0.5.1-0.20200326092940-754026bd8510
	sigs.k8s.io/controller-tools v0.2.8
	sigs.k8s.io/yaml v1.2.0
)
