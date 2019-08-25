module k8s.io/kops

go 1.12

replace k8s.io/kubernetes => k8s.io/kubernetes v1.13.5

// Version kubernetes-1.13.5
//replace k8s.io/api => k8s.io/api kubernetes-1.13.5
//replace k8s.io/apimachinery => k8s.io/apimachinery kubernetes-1.13.5
//replace k8s.io/apiserver => k8s.io/apiserver kubernetes-1.13.5
//replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver kubernetes-1.13.5
//replace k8s.io/client-go => k8s.io/client-go kubernetes-1.13.5
//replace k8s.io/cloud-provider => k8s.io/cloud-provider kubernetes-1.13.5
//replace k8s.io/utils => k8s.io/utils 66066c83e385e385ccc3c964b44fd7dcd413d0ed

							  replace k8s.io/api => k8s.io/api kubernetes-1.13.5
replace k8s.io/apimachinery => k8s.io/apimachinery kubernetes-1.13.5
replace k8s.io/apiserver => k8s.io/apiserver kubernetes-1.13.5
replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver kubernetes-1.13.5
replace k8s.io/client-go => k8s.io/client-go kubernetes-1.13.5
replace k8s.io/cloud-provider => k8s.io/cloud-provider kubernetes-1.13.5
replace k8s.io/utils => k8s.io/utils 66066c83e385e385ccc3c964b44fd7dcd413d0ed


require (
	bitbucket.org/ww/goautoneg v0.0.0-20120707110453-75cd24fc2f2c
	cloud.google.com/go v0.0.0-20160913182117-3b1ae45394a2
	github.com/Azure/go-ansiterm v0.0.0-20170929234023-d6e3b3328b78
	github.com/Azure/go-autorest v11.1.0+incompatible
	github.com/MakeNowJust/heredoc v0.0.0-20170808103936-bb23615498cd
	github.com/Masterminds/semver v1.3.1
	github.com/Masterminds/sprig v2.17.1+incompatible
	github.com/Microsoft/go-winio v0.4.5
	github.com/NYTimes/gziphandler v0.0.0-20170623195520-56545f4a5d46
	github.com/PuerkitoBio/purell v1.0.0
	github.com/PuerkitoBio/urlesc v0.0.0-20160726150825-5bd2802263f2
	github.com/aokoli/goutils v1.0.1
	github.com/aws/aws-sdk-go v1.23.0
	github.com/bazelbuild/bazel-gazelle v0.0.0-20190227183720-e443c54b396a
	github.com/bazelbuild/buildtools v0.0.0-20190213131114-55b64c3d2ddf
	github.com/beorn7/perks v0.0.0-20160229213445-3ac7bf7a47d1
	github.com/blang/semver v3.5.0+incompatible
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1
	github.com/client9/misspell v0.0.0-20170928000206-9ce5d979ffda
	github.com/coreos/etcd v3.2.24+incompatible
	github.com/coreos/go-semver v0.0.0-20150304020126-568e959cd898
	github.com/coreos/go-systemd v0.0.0-20161114122254-48702e0da86b
	github.com/cpuguy83/go-md2man v1.0.4
	github.com/davecgh/go-spew v0.0.0-20170626231645-782f4967f2dc
	github.com/denverdino/aliyungo v0.0.0-20180316152028-2581e433b270
	github.com/dgrijalva/jwt-go v0.0.0-20160705203006-01aeca54ebda
	github.com/digitalocean/godo v1.19.0
	github.com/docker/distribution v0.0.0-20170726174610-edc3ab29cdff
	github.com/docker/docker v0.0.0-20180612054059-a9fbbdc8dd87
	github.com/docker/engine-api v0.0.0-20160509170047-dea108d3aa0c
	github.com/docker/go-connections v0.3.0
	github.com/docker/go-units v0.0.0-20170127094116-9e638d38cf69
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c
	github.com/elazarl/go-bindata-assetfs v0.0.0-20150624150248-3dcc96556217
	github.com/emicklei/go-restful v0.0.0-20170410110728-ff4f55a20633
	github.com/emicklei/go-restful-swagger12 v0.0.0-20170208215640-dcef7f557305
	github.com/evanphx/json-patch v4.2.0+incompatible
	github.com/exponent-io/jsonpath v0.0.0-20151013193312-d6023ce2651d
	github.com/fatih/camelcase v0.0.0-20160318181535-f6a740d52f96
	github.com/fsnotify/fsnotify v0.0.0-20160816051541-f12c6236fe7b
	github.com/fullsailor/pkcs7 v0.0.0-20180422025557-ae226422660e
	github.com/ghodss/yaml v0.0.0-20150909031657-73d445a93680
	github.com/go-ini/ini v1.25.4
	github.com/go-openapi/analysis v0.0.0-20160815203709-b44dc874b601
	github.com/go-openapi/jsonpointer v0.0.0-20160704185906-46af16f9f7b1
	github.com/go-openapi/jsonreference v0.0.0-20160704190145-13c6e3589ad9
	github.com/go-openapi/loads v0.0.0-20170520182102-a80dea3052f0
	github.com/go-openapi/spec v0.0.0-20180213232550-1de3e0542de6
	github.com/go-openapi/swag v0.0.0-20170606142751-f3f9494671f9
	github.com/gobuffalo/envy v1.6.2
	github.com/gogo/protobuf v0.0.0-20170330071051-c0656edd0d9e
	github.com/golang/glog v0.0.0-20141105023935-44145f04b68c
	github.com/golang/groupcache v0.0.0-20160516000752-02826c3e7903
	github.com/golang/protobuf v1.1.0
	github.com/google/btree v0.0.0-20160524151835-7d79101e329e
	github.com/google/go-querystring v0.0.0-20170111101155-53e6ce116135
	github.com/google/gofuzz v0.0.0-20161122191042-44d81051d367
	github.com/google/uuid v1.1.0
	github.com/googleapis/gnostic v0.0.0-20170729233727-0c5108395e2d
	github.com/gophercloud/gophercloud v0.0.0-20190216224116-dcc6e84aef1b
	github.com/gorilla/context v1.1.1
	github.com/gorilla/mux v1.6.2
	github.com/gregjones/httpcache v0.0.0-20170728041850-787624de3eb7
	github.com/grpc-ecosystem/go-grpc-prometheus v0.0.0-20170330212424-2500245aa611
	github.com/hashicorp/golang-lru v0.0.0-20160207214719-a0d98a5f2880
	github.com/hashicorp/hcl v0.0.0-20160711231752-d8c773c4cba1
	github.com/huandu/xstrings v1.2.0
	github.com/imdario/mergo v0.3.5
	github.com/inconshreveable/mousetrap v1.0.0
	github.com/jmespath/go-jmespath v0.0.0-20160202185014-0b12d6b521d8
	github.com/joho/godotenv v1.2.0
	github.com/jpillora/backoff v0.0.0-20170918002102-8eab2debe79d
	github.com/json-iterator/go v0.0.0-20180612202835-f2b4162afba3
	github.com/jteeuwen/go-bindata v0.0.0-20151023091102-a0ff2567cfb7
	github.com/kr/fs v0.0.0-20131111012553-2788f0dbd169
	github.com/kubernetes-incubator/apiserver-builder v0.0.0-20180328231559-e809ac2f9f0c
	github.com/kubernetes-incubator/reference-docs v0.0.0-20180403034118-8fadf91876cc
	github.com/magiconair/properties v0.0.0-20160816085511-61b492c03cf4
	github.com/mailru/easyjson v0.0.0-20170624190925-2f5df55504eb
	github.com/markbates/inflect v0.0.0-20180405204719-fbc6b23ce49e
	github.com/matttproud/golang_protobuf_extensions v1.0.1
	github.com/miekg/coredns v0.0.0-20161111164017-20e25559d5ea
	github.com/miekg/dns v0.0.0-20160614162101-5d001d020961
	github.com/mitchellh/go-wordwrap v0.0.0-20150314170334-ad45545899c7
	github.com/mitchellh/mapstructure v0.0.0-20170307201123-53818660ed49
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd
	github.com/modern-go/reflect2 v0.0.0-20180701023420-4b7aa43c6742
	github.com/opencontainers/go-digest v0.0.0-20170106003457-a6d0ee40d420
	github.com/pborman/uuid v0.0.0-20150603214016-ca53cad383ca
	github.com/pelletier/go-toml v1.2.0
	github.com/petar/GoLLRB v0.0.0-20130427215148-53be0d36a84c
	github.com/peterbourgon/diskv v2.0.1+incompatible
	github.com/pkg/errors v0.8.0
	github.com/pkg/sftp v0.0.0-20160930220758-4d0e916071f6
	github.com/pmezard/go-difflib v0.0.0-20151028094244-d8ed2627bdf0
	github.com/prometheus/client_golang v0.0.0-20170531130054-e7e903064f5e
	github.com/prometheus/client_model v0.0.0-20150212101744-fa8ad6fec335
	github.com/prometheus/common v0.0.0-20170427095455-13ba4ddd0caa
	github.com/prometheus/procfs v0.0.0-20170519190837-65c1f6f8f0fc
	github.com/russross/blackfriday v0.0.0-20151117072312-300106c228d5
	github.com/sergi/go-diff v0.0.0-20161102184045-552b4e9bbdca
	github.com/shurcooL/sanitized_anchor_name v0.0.0-20151028001915-10ef21a441db
	github.com/sirupsen/logrus v0.0.0-20170822132746-89742aefa4b2
	github.com/spf13/afero v0.0.0-20160816080757-b28a7effac97
	github.com/spf13/cast v0.0.0-20160730092037-e31f36ffc91a
	github.com/spf13/cobra v0.0.0-20180319062004-c439c4fa0937
	github.com/spf13/jwalterweatherman v0.0.0-20160311093646-33c24e77fb80
	github.com/spf13/pflag v1.0.1
	github.com/spf13/viper v0.0.0-20160820190039-7fb2782df3d8
	github.com/spotinst/spotinst-sdk-go v0.0.0-20190505130751-eb52d7ac273c
	github.com/stretchr/testify v0.0.0-20180319223459-c679ae2cc0cb
	github.com/tent/http-link-go v0.0.0-20130702225549-ac974c61c2f9
	github.com/ugorji/go v0.0.0-20170107133203-ded73eae5db7
	github.com/urfave/cli v1.20.0
	github.com/vmware/govmomi v0.0.0-20180822160426-22f74650cf39
	github.com/weaveworks/mesh v0.0.0-20170419100114-1f158d31de55
	go.uber.org/atomic v1.3.2
	go.uber.org/multierr v1.1.0
	go.uber.org/zap v1.8.0
	golang.org/x/crypto v0.0.0-20180222182404-49796115aa4b
	golang.org/x/net v0.0.0-20170809000501-1c05540f6879
	golang.org/x/oauth2 v0.0.0-20170412232759-a6bd8cefa181
	golang.org/x/sys v0.0.0-20171031081856-95c657629925
	golang.org/x/text v0.0.0-20170810154203-b19bf474d317
	golang.org/x/time v0.0.0-20180412165947-fbb02b2291d2
	golang.org/x/tools v0.0.0-20170428054726-2382e3994d48
	google.golang.org/api v0.0.0-20180621000839-3639d6d93f37
	google.golang.org/appengine v1.0.0
	google.golang.org/genproto v0.0.0-20170731182057-09f6ed296fc6
	google.golang.org/grpc v1.7.5
	gopkg.in/gcfg.v1 v1.2.0
	gopkg.in/inf.v0 v0.9.0
	gopkg.in/natefinch/lumberjack.v2 v2.0.0-20150622162204-20b71e5b60d7
	gopkg.in/square/go-jose.v2 v2.0.0-20180411045311-89060dee6a84
	gopkg.in/warnings.v0 v0.1.1
	gopkg.in/yaml.v2 v2.0.0-20170721113624-670d4cfef054
	k8s.io/api v0.0.0-20190222213804-5cb15d344471
	k8s.io/apiextensions-apiserver v0.0.0-20190325193600-475668423e9f
	k8s.io/apimachinery v0.0.0-20190221213512-86fb29eff628
	k8s.io/apiserver v0.0.0-20190319190228-a4358799e4fe
	k8s.io/cli-runtime v0.0.0-20181011073557-0848ac45ae52
	k8s.io/client-go v0.0.0-20190307161346-7621a5ebb88b
	k8s.io/cloud-provider v0.0.0-20190425174118-0a4f4cbb5a66
	k8s.io/code-generator v0.0.0-20181117043124-c2090bec4d9b
	k8s.io/csi-api v0.0.0-20181011073329-55e69c84e236
	k8s.io/gengo v0.0.0-20181106084056-51747d6e00da
	k8s.io/helm v2.9.0+incompatible
	k8s.io/klog v0.3.0
	k8s.io/kube-openapi v0.0.0-20181109181836-c59034cc13d5
	k8s.io/kubernetes v1.13.5
	k8s.io/utils v0.0.0-20180726175726-66066c83e385
	sigs.k8s.io/controller-tools v0.1.10
	sigs.k8s.io/yaml v1.1.0
)
