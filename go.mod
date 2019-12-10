module k8s.io/kops

go 1.12

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
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest v11.1.2+incompatible
	github.com/GoogleCloudPlatform/k8s-cloud-provider v0.0.0-20181220005116-f8e995905100
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/semver v1.3.1
	github.com/Masterminds/sprig v2.17.1+incompatible
	github.com/Microsoft/go-winio v0.4.11
	github.com/PuerkitoBio/urlesc v0.0.0-20170810143723-de5bf2ad4578
	github.com/aokoli/goutils v1.0.1
	github.com/aws/aws-sdk-go v1.23.0
	github.com/bazelbuild/bazel-gazelle v0.18.2-0.20190823151146-67c9ddf12d8a
	github.com/bazelbuild/buildtools v0.0.0-20190731111112-f720930ceb60
	github.com/beorn7/perks v0.0.0-20180321164747-3a771d992973
	github.com/blang/semver v3.5.0+incompatible
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1
	github.com/client9/misspell v0.0.0-20170928000206-9ce5d979ffda
	github.com/coreos/etcd v3.3.13+incompatible
	github.com/coreos/go-semver v0.0.0-20180108230905-e214231b295a
	github.com/cpuguy83/go-md2man v1.0.4
	github.com/davecgh/go-spew v1.1.1
	github.com/denverdino/aliyungo v0.0.0-20180316152028-2581e433b270
	github.com/dgrijalva/jwt-go v0.0.0-20160705203006-01aeca54ebda
	github.com/digitalocean/godo v1.19.0
	github.com/docker/distribution v0.0.0-20170726174610-edc3ab29cdff
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/docker/engine-api v0.0.0-20160509170047-dea108d3aa0c
	github.com/docker/go-connections v0.3.0
	github.com/docker/go-units v0.3.3
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c
	github.com/emicklei/go-restful v0.0.0-20170410110728-ff4f55a20633
	github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d
	github.com/fatih/camelcase v0.0.0-20160318181535-f6a740d52f96
	github.com/fsnotify/fsnotify v1.4.7
	github.com/fullsailor/pkcs7 v0.0.0-20180422025557-ae226422660e
	github.com/ghodss/yaml v0.0.0-20180820084758-c7ce16629ff4
	github.com/go-ini/ini v1.25.4
	github.com/go-openapi/spec v0.17.2
	github.com/gogo/protobuf v0.0.0-20171007142547-342cbe0a0415
	github.com/golang/groupcache v0.0.0-20160516000752-02826c3e7903
	github.com/golang/protobuf v1.2.0
	github.com/google/btree v0.0.0-20160524151835-7d79101e329e
	github.com/google/go-cmp v0.3.0
	github.com/google/gofuzz v0.0.0-20170612174753-24818f796faf
	github.com/google/uuid v1.1.0
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	github.com/gophercloud/gophercloud v0.0.0-20190216224116-dcc6e84aef1b
	github.com/gorilla/mux v1.7.0
	github.com/gregjones/httpcache v0.0.0-20170728041850-787624de3eb7
	github.com/hashicorp/golang-lru v0.5.0
	github.com/hashicorp/hcl v0.0.0-20160711231752-d8c773c4cba1
	github.com/huandu/xstrings v1.2.0
	github.com/imdario/mergo v0.3.5
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/jmespath/go-jmespath v0.0.0-20180206201540-c2b33e8439af
	github.com/jpillora/backoff v0.0.0-20170918002102-8eab2debe79d
	github.com/json-iterator/go v0.0.0-20180701071628-ab8a2e0c74be
	github.com/jteeuwen/go-bindata v0.0.0-20151023091102-a0ff2567cfb7
	github.com/kr/fs v0.0.0-20131111012553-2788f0dbd169
	github.com/magiconair/properties v0.0.0-20160816085511-61b492c03cf4
	github.com/mailru/easyjson v0.0.0-20180823135443-60711f1a8329
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/miekg/coredns v0.0.0-20161111164017-20e25559d5ea
	github.com/miekg/dns v0.0.0-20160614162101-5d001d020961
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7
	github.com/mitchellh/mapstructure v1.1.2
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v1.0.1
	github.com/opencontainers/go-digest v0.0.0-20170106003457-a6d0ee40d420
	github.com/pborman/uuid v1.2.0
	github.com/pelletier/go-toml v1.2.0
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors v0.8.0
	github.com/pkg/sftp v0.0.0-20160930220758-4d0e916071f6
	github.com/pmezard/go-difflib v1.0.0
	github.com/prometheus/client_golang v0.9.2
	github.com/prometheus/client_model v0.0.0-20180712105110-5c3871d89910
	github.com/prometheus/common v0.0.0-20181126121408-4724e9255275
	github.com/prometheus/procfs v0.0.0-20181204211112-1dc9a6cbc91a
	github.com/russross/blackfriday v0.0.0-20151117072312-300106c228d5
	github.com/sergi/go-diff v0.0.0-20161102184045-552b4e9bbdca
	github.com/shurcooL/sanitized_anchor_name v0.0.0-20151028001915-10ef21a441db
	github.com/sirupsen/logrus v1.2.0
	github.com/spf13/afero v0.0.0-20160816080757-b28a7effac97
	github.com/spf13/cast v0.0.0-20160730092037-e31f36ffc91a
	github.com/spf13/cobra v0.0.0-20180319062004-c439c4fa0937
	github.com/spf13/jwalterweatherman v0.0.0-20160311093646-33c24e77fb80
	github.com/spf13/pflag v1.0.1
	github.com/spf13/viper v0.0.0-20160820190039-7fb2782df3d8
	github.com/spotinst/spotinst-sdk-go v0.0.0-20190505130751-eb52d7ac273c
	github.com/stretchr/testify v1.3.0
	github.com/urfave/cli v1.20.0
	github.com/vmware/govmomi v0.20.1
	github.com/weaveworks/mesh v0.0.0-20170419100114-1f158d31de55
	go.uber.org/atomic v1.3.2
	go.uber.org/multierr v1.1.0
	go.uber.org/zap v1.8.0
	golang.org/x/crypto v0.0.0-20190308221718-c2843e01d9a2
	golang.org/x/net v0.0.0-20190812203447-cdfb69ac37fc
	golang.org/x/oauth2 v0.0.0-20190402181905-9f3314589c9a
	golang.org/x/sys v0.0.0-20190312061237-fead79001313
	golang.org/x/text v0.3.1-0.20181227161524-e6919f6577db
	golang.org/x/time v0.0.0-20180412165947-fbb02b2291d2
	golang.org/x/tools v0.0.0-20190328211700-ab21143f2384
	google.golang.org/api v0.0.0-20181220000619-583d854617af
	gopkg.in/gcfg.v1 v1.2.0
	gopkg.in/inf.v0 v0.9.0
	gopkg.in/warnings.v0 v0.1.1
	gopkg.in/yaml.v2 v2.2.2
	k8s.io/api v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/cli-runtime v0.0.0
	k8s.io/client-go v0.0.0
	k8s.io/cloud-provider v0.0.0
	k8s.io/csi-translation-lib v0.0.0
	k8s.io/helm v2.9.0+incompatible
	k8s.io/klog v0.3.1
	k8s.io/kube-openapi v0.0.0-20190228160746-b3a7cee44a30
	k8s.io/kubernetes v1.15.3
	k8s.io/legacy-cloud-providers v0.0.0
	k8s.io/utils v0.0.0-20190221042446-c2654d5206da
	sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml v1.1.0
)
