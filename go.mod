module k8s.io/kops

go 1.13

// Version kubernetes-1.15.3
//replace k8s.io/kubernetes => k8s.io/kubernetes v1.15.3
//replace k8s.io/api => k8s.io/api kubernetes-1.15.3
//replace k8s.io/apimachinery => k8s.io/apimachinery kubernetes-1.15.3
//replace k8s.io/client-go => k8s.io/client-go kubernetes-1.15.3
//replace k8s.io/cloud-provider => k8s.io/cloud-provider kubernetes-1.15.3
//replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers kubernetes-1.15.3

replace k8s.io/kubernetes => k8s.io/kubernetes v1.15.3

replace k8s.io/api => k8s.io/api v0.0.0-20190819141258-3544db3b9e44

replace k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190817020851-f2f3a405f61d

replace k8s.io/client-go => k8s.io/client-go v0.0.0-20190819141724-e14f31a72a77

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190819145148-d91c85d212d5

// Dependencies we don't really need, except that kubernetes specifies them as v0.0.0 which confuses go.mod
//replace k8s.io/apiserver => k8s.io/apiserver kubernetes-1.15.3
//replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver kubernetes-1.15.3
//replace k8s.io/kube-scheduler => k8s.io/kube-scheduler kubernetes-1.15.3
//replace k8s.io/kube-proxy => k8s.io/kube-proxy kubernetes-1.15.3
//replace k8s.io/cri-api => k8s.io/cri-api kubernetes-1.15.3
//replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib kubernetes-1.15.3
//replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers kubernetes-1.15.3
//replace k8s.io/component-base => k8s.io/component-base kubernetes-1.15.3
//replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap kubernetes-1.15.3
//replace k8s.io/metrics => k8s.io/metrics kubernetes-1.15.3
//replace k8s.io/sample-apiserver => k8s.io/sample-apiserver kubernetes-1.15.3
//replace k8s.io/kube-aggregator => k8s.io/kube-aggregator kubernetes-1.15.3
//replace k8s.io/kubelet => k8s.io/kubelet kubernetes-1.15.3
//replace k8s.io/cli-runtime => k8s.io/cli-runtime kubernetes-1.15.3
//replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager kubernetes-1.15.3
//replace k8s.io/code-generator => k8s.io/code-generator kubernetes-1.15.3

replace k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190819142446-92cc630367d0

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190819143637-0dbe462fe92d

replace k8s.io/kubelet => k8s.io/kubelet v0.0.0-20190819144524-827174bad5e8

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190819144027-541433d7ce35

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.0.0-20190819144832-f53437941eef

replace k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190612205613-18da4a14b22b

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.0.0-20190819144657-d1a724e0828e

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.0.0-20190819144346-2e47de1df0f0

replace k8s.io/cri-api => k8s.io/cri-api v0.0.0-20190817025403-3ae76f584e79

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.0.0-20190819145328-4831a4ced492

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.0.0-20190819145509-592c9a46fd00

replace k8s.io/component-base => k8s.io/component-base v0.0.0-20190819141909-f0f7c184477d

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.0.0-20190819145008-029dd04813af

replace k8s.io/metrics => k8s.io/metrics v0.0.0-20190819143841-305e1cef1ab1

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.0.0-20190819143045-c84c31c165c4

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190819142756-13daafd3604f

require (
	cloud.google.com/go v0.34.0
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/semver v1.3.1 // indirect
	github.com/Masterminds/sprig v2.17.1+incompatible
	github.com/aokoli/goutils v1.0.1 // indirect
	github.com/aws/aws-sdk-go v1.30.16
	github.com/bazelbuild/bazel-gazelle v0.19.1
	github.com/blang/semver v3.5.0+incompatible
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/client9/misspell v0.0.0-20170928000206-9ce5d979ffda
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/denverdino/aliyungo v0.0.0-20180316152028-2581e433b270
	github.com/digitalocean/godo v1.19.0
	github.com/docker/engine-api v0.0.0-20160509170047-dea108d3aa0c
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/fullsailor/pkcs7 v0.0.0-20180422025557-ae226422660e
	github.com/ghodss/yaml v0.0.0-20180820084758-c7ce16629ff4
	github.com/go-ini/ini v1.51.0
	github.com/go-logr/logr v0.1.0
	github.com/gogo/protobuf v1.1.1
	github.com/golang/protobuf v1.3.2
	github.com/google/uuid v1.1.0 // indirect
	github.com/gophercloud/gophercloud v0.7.1-0.20200116011225-46fdd1830e9a
	github.com/gorilla/mux v1.7.0
	github.com/hashicorp/hcl v1.0.0
	github.com/huandu/xstrings v1.2.0 // indirect
	github.com/jacksontj/memberlistmesh v0.0.0-20190905163944-93462b9d2bb7
	github.com/jpillora/backoff v0.0.0-20170918002102-8eab2debe79d
	github.com/jteeuwen/go-bindata v0.0.0-20151023091102-a0ff2567cfb7
	github.com/miekg/coredns v0.0.0-20161111164017-20e25559d5ea
	github.com/miekg/dns v1.0.14
	github.com/mitchellh/mapstructure v1.1.2
	github.com/pborman/uuid v1.2.0
	github.com/pkg/errors v0.9.1
	github.com/pkg/sftp v0.0.0-20160930220758-4d0e916071f6
	github.com/prometheus/client_golang v0.9.2
	github.com/sergi/go-diff v0.0.0-20161102184045-552b4e9bbdca
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.3
	github.com/spf13/viper v1.3.2
	github.com/spotinst/spotinst-sdk-go v1.36.1
	github.com/stretchr/testify v1.5.1
	github.com/urfave/cli v1.20.0
	github.com/vmware/govmomi v0.20.1
	github.com/weaveworks/mesh v0.0.0-20170419100114-1f158d31de55
	go.uber.org/zap v1.9.1
	golang.org/x/crypto v0.0.0-20191202143827-86a70503ff7e
	golang.org/x/net v0.0.0-20200202094626-16171245cfb2
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sys v0.0.0-20191128015809-6d18c012aee9
	google.golang.org/api v0.0.0-20181220000619-583d854617af
	gopkg.in/gcfg.v1 v1.2.0
	gopkg.in/inf.v0 v0.9.1
	gopkg.in/yaml.v2 v2.2.7
	honnef.co/go/tools v0.0.1-2019.2.3
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/helm v2.9.0+incompatible
	k8s.io/klog v0.3.1
	k8s.io/kubernetes v1.15.3
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/utils v0.0.0-20190506122338-8fab8cb257d5
	sigs.k8s.io/controller-runtime v0.2.2
	sigs.k8s.io/controller-tools v0.2.2-0.20190919191502-76a25b63325a
	sigs.k8s.io/yaml v1.1.0
)
