module k8s.io/kops/tests/e2e

go 1.17

require (
	github.com/blang/semver v3.5.1+incompatible
	github.com/google/shlex v0.0.0-20191202100458-e7afc7fbc510
	github.com/octago/sflags v0.2.0
	github.com/spf13/pflag v1.0.5
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.23.5
	k8s.io/apimachinery v0.23.5
	k8s.io/client-go v9.0.0+incompatible
	k8s.io/klog/v2 v2.30.0
	k8s.io/kops v0.0.0-00010101000000-000000000000
	sigs.k8s.io/boskos v0.0.0-20200710214748-f5935686c7fc
	sigs.k8s.io/kubetest2 v0.0.0-20220224035534-5e5d3e9eebc6
	sigs.k8s.io/yaml v1.3.0
)

require (
	cloud.google.com/go v0.99.0 // indirect
	cloud.google.com/go/storage v1.12.0 // indirect
	github.com/Azure/azure-pipeline-go v0.2.3 // indirect
	github.com/Azure/azure-storage-blob-go v0.13.0 // indirect
	github.com/StackExchange/wmi v0.0.0-20190523213315-cbe66965904d // indirect
	github.com/apparentlymart/go-cidr v1.1.0 // indirect
	github.com/aws/aws-sdk-go v1.43.11 // indirect
	github.com/blang/semver/v4 v4.0.0 // indirect
	github.com/cenkalti/backoff/v3 v3.0.0 // indirect
	github.com/containers/image/v5 v5.9.0 // indirect
	github.com/containers/libtrust v0.0.0-20190913040956-14b96171aa3b // indirect
	github.com/containers/ocicrypt v1.1.1 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/docker/cli v20.10.10+incompatible // indirect
	github.com/docker/distribution v2.7.1+incompatible // indirect
	github.com/docker/docker v20.10.10+incompatible // indirect
	github.com/docker/docker-credential-helpers v0.6.4 // indirect
	github.com/emirpasic/gods v1.12.0 // indirect
	github.com/go-git/gcfg v1.5.0 // indirect
	github.com/go-git/go-billy/v5 v5.0.0 // indirect
	github.com/go-git/go-git/v5 v5.2.0 // indirect
	github.com/go-ini/ini v1.62.0 // indirect
	github.com/go-logr/logr v1.2.0 // indirect
	github.com/go-ole/go-ole v1.2.4 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/golang/snappy v0.0.3 // indirect
	github.com/google/go-cmp v0.5.6 // indirect
	github.com/google/go-containerregistry v0.7.0 // indirect
	github.com/google/go-github/v33 v33.0.0 // indirect
	github.com/google/go-querystring v1.0.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/googleapis/gax-go/v2 v2.1.1 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/gophercloud/gophercloud v0.24.0 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-multierror v1.1.0 // indirect
	github.com/hashicorp/go-retryablehttp v0.6.6 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/go-sockaddr v1.0.2 // indirect
	github.com/hashicorp/hcl v1.0.0 // indirect
	github.com/hashicorp/vault/api v1.1.1 // indirect
	github.com/hashicorp/vault/sdk v0.2.1 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/jbenet/go-context v0.0.0-20150711004518-d14ea06fba99 // indirect
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/kballard/go-shellquote v0.0.0-20180428030007-95032a82bc51 // indirect
	github.com/kevinburke/ssh_config v0.0.0-20190725054713-01f96b0aa0cd // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/klauspost/pgzip v1.2.5 // indirect
	github.com/kr/fs v0.1.0 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/mattn/go-ieproxy v0.0.1 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.4.3 // indirect
	github.com/moby/spdystream v0.2.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/opencontainers/go-digest v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.2-0.20210730191737-8e42a01fb1b7 // indirect
	github.com/pierrec/lz4 v2.5.2+incompatible // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pkg/sftp v1.13.4 // indirect
	github.com/ryanuber/go-glob v1.0.0 // indirect
	github.com/sergi/go-diff v1.2.0 // indirect
	github.com/shirou/gopsutil/v3 v3.20.12 // indirect
	github.com/sirupsen/logrus v1.8.1 // indirect
	github.com/smartystreets/assertions v1.2.0 // indirect
	github.com/spf13/cobra v1.3.0 // indirect
	github.com/ulikunitz/xz v0.5.8 // indirect
	github.com/xanzy/ssh-agent v0.2.1 // indirect
	go.opencensus.io v0.23.0 // indirect
	golang.org/x/crypto v0.0.0-20211215153901-e495a2d5b3d3 // indirect
	golang.org/x/net v0.0.0-20220127200216-cd36cc0744dd // indirect
	golang.org/x/oauth2 v0.0.0-20211104180415-d3ed0bb246c8 // indirect
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
	golang.org/x/term v0.0.0-20210927222741-03fcf44c2211 // indirect
	golang.org/x/text v0.3.7 // indirect
	golang.org/x/time v0.0.0-20210723032227-1f47c861a9ac // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/api v0.62.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20211208223120-3a66f561d7aa // indirect
	google.golang.org/grpc v1.42.0 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
	gopkg.in/square/go-jose.v2 v2.5.1 // indirect
	gopkg.in/warnings.v0 v0.1.2 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	k8s.io/klog v1.0.0 // indirect
	k8s.io/release v0.7.1-0.20210204090829-09fb5e3883b8 // indirect
	k8s.io/test-infra v0.0.0-20200617221206-ea73eaeab7ff // indirect
	k8s.io/utils v0.0.0-20211116205334-6203023598ed // indirect
	sigs.k8s.io/structured-merge-diff/v4 v4.2.1 // indirect
)

replace k8s.io/kops => ../../.

// These should match the go.mod from k8s.io/kops

replace k8s.io/api => k8s.io/api v0.22.2

replace k8s.io/apimachinery => k8s.io/apimachinery v0.22.2

replace k8s.io/client-go => k8s.io/client-go v0.22.2

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.22.2

replace k8s.io/controller-manager => k8s.io/controller-manager v0.22.2

replace k8s.io/kubectl => k8s.io/kubectl v0.22.2

replace k8s.io/apiserver => k8s.io/apiserver v0.22.2

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.22.2

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.22.2

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.22.2

replace k8s.io/cri-api => k8s.io/cri-api v0.22.2

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.22.2

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.22.2

replace k8s.io/component-base => k8s.io/component-base v0.22.2

replace k8s.io/component-helpers => k8s.io/component-helpers v0.22.2

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.22.2

replace k8s.io/metrics => k8s.io/metrics v0.22.2

replace k8s.io/mount-utils => k8s.io/mount-utils v0.22.2

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.22.2

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.22.2

replace k8s.io/kubelet => k8s.io/kubelet v0.22.2

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.2

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.22.2
