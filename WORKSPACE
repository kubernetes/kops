http_archive(
    name = "io_bazel_rules_go",
    url = "https://github.com/bazelbuild/rules_go/releases/download/0.8.1/rules_go-0.8.1.tar.gz",
    sha256 = "90bb270d0a92ed5c83558b2797346917c46547f6f7103e648941ecdb6b9d0e72",
)

load("@io_bazel_rules_go//go:def.bzl", "go_register_toolchains", "go_repository", "go_rules_dependencies")

go_rules_dependencies()

go_register_toolchains()

load("@io_bazel_rules_go//proto:def.bzl", "proto_register_toolchains")

proto_register_toolchains()

http_archive(
    name = "bazel_gazelle",
    url = "https://github.com/bazelbuild/bazel-gazelle/releases/download/0.8/bazel-gazelle-0.8.tar.gz",
    sha256 = "e3dadf036c769d1f40603b86ae1f0f90d11837116022d9b06e4cd88cae786676",
)

load("@bazel_gazelle//:deps.bzl", "gazelle_dependencies")

gazelle_dependencies()

http_archive(
    name = "io_kubernetes_build",
    sha256 = "cf138e48871629345548b4aaf23101314b5621c1bdbe45c4e75edb45b08891f0",
    strip_prefix = "repo-infra-1fb0a3ff0cc5308a6d8e2f3f9c57d1f2f940354e",
    urls = ["https://github.com/kubernetes/repo-infra/archive/1fb0a3ff0cc5308a6d8e2f3f9c57d1f2f940354e.tar.gz"],
)

go_repository(
    name = "com_github_aokoli_goutils",
    commit = "3391d3790d23d03408670993e957e8f408993c34",
    importpath = "github.com/aokoli/goutils",
)

go_repository(
    name = "com_github_aws_aws_sdk_go",
    commit = "5b4d64c21c9af31b5c870528001cd2e3ef37bfe0",
    importpath = "github.com/aws/aws-sdk-go",
)

go_repository(
    name = "com_github_azure_go_ansiterm",
    commit = "d6e3b3328b783f23731bc4d058875b0371ff8109",
    importpath = "github.com/Azure/go-ansiterm",
)

go_repository(
    name = "com_github_azure_go_autorest",
    commit = "809ed2ef5c4c9a60c3c2f3aa9cc11f3a7c2ce59d",
    importpath = "github.com/Azure/go-autorest",
)

go_repository(
    name = "com_github_beorn7_perks",
    commit = "4c0e84591b9aa9e6dcfdf3e020114cd81f89d5f9",
    importpath = "github.com/beorn7/perks",
)

go_repository(
    name = "com_github_blang_semver",
    commit = "2ee87856327ba09384cabd113bc6b5d174e9ec0f",
    importpath = "github.com/blang/semver",
)

go_repository(
    name = "com_github_chai2010_gettext_go",
    commit = "bf70f2a70fb1b1f36d90d671a72795984eab0fcb",
    importpath = "github.com/chai2010/gettext-go",
)

go_repository(
    name = "com_github_coredns_coredns",
    commit = "1b60688dc8f7e7ed0edcda5fb3fec02b1ddd3e42",
    importpath = "github.com/coredns/coredns",
)

go_repository(
    name = "com_github_coreos_etcd",
    commit = "1e1dbb23924672c6cd72c62ee0db2b45f778da71",
    importpath = "github.com/coreos/etcd",
)

go_repository(
    name = "com_github_coreos_go_semver",
    commit = "8ab6407b697782a06568d4b7f1db25550ec2e4c6",
    importpath = "github.com/coreos/go-semver",
)

go_repository(
    name = "com_github_coreos_go_systemd",
    commit = "d2196463941895ee908e13531a23a39feb9e1243",
    importpath = "github.com/coreos/go-systemd",
)

go_repository(
    name = "com_github_cpuguy83_go_md2man",
    commit = "71acacd42f85e5e82f70a55327789582a5200a90",
    importpath = "github.com/cpuguy83/go-md2man",
)

go_repository(
    name = "com_github_davecgh_go_spew",
    commit = "346938d642f2ec3594ed81d874461961cd0faa76",
    importpath = "github.com/davecgh/go-spew",
)

go_repository(
    name = "com_github_daviddengcn_go_colortext",
    commit = "17e75f6184bc9e727756cd0d82e0af58b1fc7191",
    importpath = "github.com/daviddengcn/go-colortext",
)

go_repository(
    name = "com_github_dgrijalva_jwt_go",
    commit = "dbeaa9332f19a944acb5736b4456cfcc02140e29",
    importpath = "github.com/dgrijalva/jwt-go",
)

go_repository(
    name = "com_github_digitalocean_godo",
    commit = "77ea48de76a7b31b234d854f15d003c68bb2fb90",
    importpath = "github.com/digitalocean/godo",
)

go_repository(
    name = "com_github_docker_distribution",
    commit = "edc3ab29cdff8694dd6feb85cfeb4b5f1b38ed9c",
    importpath = "github.com/docker/distribution",
)

go_repository(
    name = "com_github_docker_docker",
    commit = "092cba3727bb9b4a2f0e922cd6c0f93ea270e363",
    importpath = "github.com/docker/docker",
)

go_repository(
    name = "com_github_docker_engine_api",
    commit = "dea108d3aa0c67d7162a3fd8aa65f38a430019fd",
    importpath = "github.com/docker/engine-api",
)

go_repository(
    name = "com_github_docker_go_connections",
    commit = "3ede32e2033de7505e6500d6c868c2b9ed9f169d",
    importpath = "github.com/docker/go-connections",
)

go_repository(
    name = "com_github_docker_go_units",
    commit = "0dadbb0345b35ec7ef35e228dabb8de89a65bf52",
    importpath = "github.com/docker/go-units",
)

go_repository(
    name = "com_github_docker_spdystream",
    commit = "bc6354cbbc295e925e4c611ffe90c1f287ee54db",
    importpath = "github.com/docker/spdystream",
)

go_repository(
    name = "com_github_elazarl_go_bindata_assetfs",
    commit = "30f82fa23fd844bd5bb1e5f216db87fd77b5eb43",
    importpath = "github.com/elazarl/go-bindata-assetfs",
)

go_repository(
    name = "com_github_emicklei_go_restful",
    commit = "5741799b275a3c4a5a9623a993576d7545cf7b5c",
    importpath = "github.com/emicklei/go-restful",
)

go_repository(
    name = "com_github_emicklei_go_restful_swagger12",
    commit = "dcef7f55730566d41eae5db10e7d6981829720f6",
    importpath = "github.com/emicklei/go-restful-swagger12",
)

go_repository(
    name = "com_github_evanphx_json_patch",
    commit = "944e07253867aacae43c04b2e6a239005443f33a",
    importpath = "github.com/evanphx/json-patch",
)

go_repository(
    name = "com_github_exponent_io_jsonpath",
    commit = "d6023ce2651d8eafb5c75bb0c7167536102ec9f5",
    importpath = "github.com/exponent-io/jsonpath",
)

go_repository(
    name = "com_github_fatih_camelcase",
    commit = "44e46d280b43ec1531bb25252440e34f1b800b65",
    importpath = "github.com/fatih/camelcase",
)

go_repository(
    name = "com_github_fsnotify_fsnotify",
    commit = "629574ca2a5df945712d3079857300b5e4da0236",
    importpath = "github.com/fsnotify/fsnotify",
)

go_repository(
    name = "com_github_ghodss_yaml",
    commit = "0ca9ea5df5451ffdf184b4428c902747c2c11cd7",
    importpath = "github.com/ghodss/yaml",
)

go_repository(
    name = "com_github_go_ini_ini",
    commit = "32e4c1e6bc4e7d0d8451aa6b75200d19e37a536a",
    importpath = "github.com/go-ini/ini",
)

go_repository(
    name = "com_github_go_openapi_jsonpointer",
    commit = "779f45308c19820f1a69e9a4cd965f496e0da10f",
    importpath = "github.com/go-openapi/jsonpointer",
)

go_repository(
    name = "com_github_go_openapi_jsonreference",
    commit = "36d33bfe519efae5632669801b180bf1a245da3b",
    importpath = "github.com/go-openapi/jsonreference",
)

go_repository(
    name = "com_github_go_openapi_spec",
    commit = "01738944bdee0f26bf66420c5b17d54cfdf55341",
    importpath = "github.com/go-openapi/spec",
)

go_repository(
    name = "com_github_go_openapi_swag",
    commit = "cf0bdb963811675a4d7e74901cefc7411a1df939",
    importpath = "github.com/go-openapi/swag",
)

go_repository(
    name = "com_github_gogo_protobuf",
    commit = "342cbe0a04158f6dcb03ca0079991a51a4248c02",
    importpath = "github.com/gogo/protobuf",
)

go_repository(
    name = "com_github_golang_glog",
    commit = "23def4e6c14b4da8ac2ed8007337bc5eb5007998",
    importpath = "github.com/golang/glog",
)

go_repository(
    name = "com_github_golang_groupcache",
    commit = "84a468cf14b4376def5d68c722b139b881c450a4",
    importpath = "github.com/golang/groupcache",
)

go_repository(
    name = "com_github_golang_protobuf",
    commit = "1e59b77b52bf8e4b449a57e6f79f21226d571845",
    importpath = "github.com/golang/protobuf",
)

go_repository(
    name = "com_github_google_btree",
    commit = "316fb6d3f031ae8f4d457c6c5186b9e3ded70435",
    importpath = "github.com/google/btree",
)

go_repository(
    name = "com_github_google_cadvisor",
    commit = "1e567c2ac359c3ed1303e0c80b6cf08edefc841d",
    importpath = "github.com/google/cadvisor",
)

go_repository(
    name = "com_github_google_go_querystring",
    commit = "53e6ce116135b80d037921a7fdd5138cf32d7a8a",
    importpath = "github.com/google/go-querystring",
)

go_repository(
    name = "com_github_google_gofuzz",
    commit = "24818f796faf91cd76ec7bddd72458fbced7a6c1",
    importpath = "github.com/google/gofuzz",
)

go_repository(
    name = "com_github_googleapis_gnostic",
    commit = "ee43cbb60db7bd22502942cccbc39059117352ab",
    importpath = "github.com/googleapis/gnostic",
)

go_repository(
    name = "com_github_gophercloud_gophercloud",
    commit = "caa74f7b5b95aa1ed9d52e51e5c097bef4e26e52",
    importpath = "github.com/gophercloud/gophercloud",
)

go_repository(
    name = "com_github_gregjones_httpcache",
    commit = "2bcd89a1743fd4b373f7370ce8ddc14dfbd18229",
    importpath = "github.com/gregjones/httpcache",
)

go_repository(
    name = "com_github_hashicorp_golang_lru",
    commit = "0a025b7e63adc15a622f29b0b2c4c3848243bbf6",
    importpath = "github.com/hashicorp/golang-lru",
)

go_repository(
    name = "com_github_hashicorp_hcl",
    commit = "23c074d0eceb2b8a5bfdbb271ab780cde70f05a8",
    importpath = "github.com/hashicorp/hcl",
)

go_repository(
    name = "com_github_howeyc_gopass",
    commit = "bf9dde6d0d2c004a008c27aaee91170c786f6db8",
    importpath = "github.com/howeyc/gopass",
)

go_repository(
    name = "com_github_huandu_xstrings",
    commit = "37469d0c81a7910b49d64a0d308ded4823e90937",
    importpath = "github.com/huandu/xstrings",
)

go_repository(
    name = "com_github_imdario_mergo",
    commit = "7fe0c75c13abdee74b09fcacef5ea1c6bba6a874",
    importpath = "github.com/imdario/mergo",
)

go_repository(
    name = "com_github_inconshreveable_mousetrap",
    commit = "76626ae9c91c4f2a10f34cad8ce83ea42c93bb75",
    importpath = "github.com/inconshreveable/mousetrap",
)

go_repository(
    name = "com_github_jmespath_go_jmespath",
    commit = "3433f3ea46d9f8019119e7dd41274e112a2359a9",
    importpath = "github.com/jmespath/go-jmespath",
)

go_repository(
    name = "com_github_jonboulle_clockwork",
    commit = "2eee05ed794112d45db504eb05aa693efd2b8b09",
    importpath = "github.com/jonboulle/clockwork",
)

go_repository(
    name = "com_github_json_iterator_go",
    commit = "f7279a603edee96fe7764d3de9c6ff8cf9970994",
    importpath = "github.com/json-iterator/go",
)

go_repository(
    name = "com_github_juju_ratelimit",
    commit = "59fac5042749a5afb9af70e813da1dd5474f0167",
    importpath = "github.com/juju/ratelimit",
)

go_repository(
    name = "com_github_kr_fs",
    commit = "2788f0dbd16903de03cb8186e5c7d97b69ad387b",
    importpath = "github.com/kr/fs",
)

go_repository(
    name = "com_github_magiconair_properties",
    commit = "d419a98cdbed11a922bf76f257b7c4be79b50e73",
    importpath = "github.com/magiconair/properties",
)

go_repository(
    name = "com_github_mailru_easyjson",
    commit = "32fa128f234d041f196a9f3e0fea5ac9772c08e1",
    importpath = "github.com/mailru/easyjson",
)

go_repository(
    name = "com_github_makenowjust_heredoc",
    commit = "e9091a26100e9cfb2b6a8f470085bfa541931a91",
    importpath = "github.com/MakeNowJust/heredoc",
)

go_repository(
    name = "com_github_masterminds_semver",
    commit = "15d8430ab86497c5c0da827b748823945e1cf1e1",
    importpath = "github.com/Masterminds/semver",
)

go_repository(
    name = "com_github_masterminds_sprig",
    commit = "b217b9c388de2cacde4354c536e520c52c055563",
    importpath = "github.com/Masterminds/sprig",
)

go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    commit = "3247c84500bff8d9fb6d579d800f20b3e091582c",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
)

go_repository(
    name = "com_github_microsoft_go_winio",
    commit = "78439966b38d69bf38227fbf57ac8a6fee70f69a",
    importpath = "github.com/Microsoft/go-winio",
)

go_repository(
    name = "com_github_miekg_coredns",
    commit = "1b60688dc8f7e7ed0edcda5fb3fec02b1ddd3e42",
    importpath = "github.com/miekg/coredns",
)

go_repository(
    name = "com_github_miekg_dns",
    commit = "9271f6595be6734763aaf4b7923f8f1b7a6cc559",
    importpath = "github.com/miekg/dns",
)

go_repository(
    name = "com_github_mitchellh_go_wordwrap",
    commit = "ad45545899c7b13c020ea92b2072220eefad42b8",
    importpath = "github.com/mitchellh/go-wordwrap",
)

go_repository(
    name = "com_github_mitchellh_mapstructure",
    commit = "06020f85339e21b2478f756a78e295255ffa4d6a",
    importpath = "github.com/mitchellh/mapstructure",
)

go_repository(
    name = "com_github_mxk_go_flowrate",
    commit = "cca7078d478f8520f85629ad7c68962d31ed7682",
    importpath = "github.com/mxk/go-flowrate",
)

go_repository(
    name = "com_github_nytimes_gziphandler",
    commit = "d6f46609c7629af3a02d791a4666866eed3cbd3e",
    importpath = "github.com/NYTimes/gziphandler",
)

go_repository(
    name = "com_github_opencontainers_go_digest",
    commit = "279bed98673dd5bef374d3b6e4b09e2af76183bf",
    importpath = "github.com/opencontainers/go-digest",
)

go_repository(
    name = "com_github_pborman_uuid",
    commit = "e790cca94e6cc75c7064b1332e63811d4aae1a53",
    importpath = "github.com/pborman/uuid",
)

go_repository(
    name = "com_github_pelletier_go_toml",
    commit = "16398bac157da96aa88f98a2df640c7f32af1da2",
    importpath = "github.com/pelletier/go-toml",
)

go_repository(
    name = "com_github_petar_gollrb",
    commit = "53be0d36a84c2a886ca057d34b6aa4468df9ccb4",
    importpath = "github.com/petar/GoLLRB",
)

go_repository(
    name = "com_github_peterbourgon_diskv",
    commit = "5f041e8faa004a95c88a202771f4cc3e991971e6",
    importpath = "github.com/peterbourgon/diskv",
)

go_repository(
    name = "com_github_pkg_errors",
    commit = "645ef00459ed84a119197bfb8d8205042c6df63d",
    importpath = "github.com/pkg/errors",
)

go_repository(
    name = "com_github_pkg_sftp",
    commit = "1d5374a61d4959af383169ff31db1cd752c2d69a",
    importpath = "github.com/pkg/sftp",
)

go_repository(
    name = "com_github_prometheus_client_golang",
    commit = "c5b7fccd204277076155f10851dad72b76a49317",
    importpath = "github.com/prometheus/client_golang",
)

go_repository(
    name = "com_github_prometheus_client_model",
    commit = "99fa1f4be8e564e8a6b613da7fa6f46c9edafc6c",
    importpath = "github.com/prometheus/client_model",
)

go_repository(
    name = "com_github_prometheus_common",
    commit = "2e54d0b93cba2fd133edc32211dcc32c06ef72ca",
    importpath = "github.com/prometheus/common",
)

go_repository(
    name = "com_github_prometheus_procfs",
    commit = "f98634e408857669d61064b283c4cde240622865",
    importpath = "github.com/prometheus/procfs",
)

go_repository(
    name = "com_github_puerkitobio_purell",
    commit = "0bcb03f4b4d0a9428594752bd2a3b9aa0a9d4bd4",
    importpath = "github.com/PuerkitoBio/purell",
)

go_repository(
    name = "com_github_puerkitobio_urlesc",
    commit = "de5bf2ad457846296e2031421a34e2568e304e35",
    importpath = "github.com/PuerkitoBio/urlesc",
)

go_repository(
    name = "com_github_renstrom_dedent",
    commit = "a1eba44eaecc89804e4b05ce2d17168eb353d524",
    importpath = "github.com/renstrom/dedent",
)

go_repository(
    name = "com_github_russross_blackfriday",
    commit = "300106c228d52c8941d4b3de6054a6062a86dda3",
    importpath = "github.com/russross/blackfriday",
)

go_repository(
    name = "com_github_satori_go_uuid",
    commit = "879c5887cd475cd7864858769793b2ceb0d44feb",
    importpath = "github.com/satori/go.uuid",
)

go_repository(
    name = "com_github_sergi_go_diff",
    commit = "1744e2970ca51c86172c8190fadad617561ed6e7",
    importpath = "github.com/sergi/go-diff",
)

go_repository(
    name = "com_github_shurcool_sanitized_anchor_name",
    commit = "86672fcb3f950f35f2e675df2240550f2a50762f",
    importpath = "github.com/shurcooL/sanitized_anchor_name",
)

go_repository(
    name = "com_github_sirupsen_logrus",
    commit = "d682213848ed68c0a260ca37d6dd5ace8423f5ba",
    importpath = "github.com/Sirupsen/logrus",
)

go_repository(
    name = "com_github_spf13_afero",
    commit = "8d919cbe7e2627e417f3e45c3c0e489a5b7e2536",
    importpath = "github.com/spf13/afero",
)

go_repository(
    name = "com_github_spf13_cast",
    commit = "acbeb36b902d72a7a4c18e8f3241075e7ab763e4",
    importpath = "github.com/spf13/cast",
)

go_repository(
    name = "com_github_spf13_cobra",
    commit = "7b2c5ac9fc04fc5efafb60700713d4fa609b777b",
    importpath = "github.com/spf13/cobra",
)

go_repository(
    name = "com_github_spf13_jwalterweatherman",
    commit = "12bd96e66386c1960ab0f74ced1362f66f552f7b",
    importpath = "github.com/spf13/jwalterweatherman",
)

go_repository(
    name = "com_github_spf13_pflag",
    commit = "e57e3eeb33f795204c1ca35f56c44f83227c6e66",
    importpath = "github.com/spf13/pflag",
)

go_repository(
    name = "com_github_spf13_viper",
    commit = "25b30aa063fc18e48662b86996252eabdcf2f0c7",
    importpath = "github.com/spf13/viper",
)

go_repository(
    name = "com_github_tent_http_link_go",
    commit = "ac974c61c2f990f4115b119354b5e0b47550e888",
    importpath = "github.com/tent/http-link-go",
)

go_repository(
    name = "com_github_ugorji_go",
    commit = "ded73eae5db7e7a0ef6f55aace87a2873c5d2b74",
    importpath = "github.com/ugorji/go",
)

go_repository(
    name = "com_github_vmware_govmomi",
    commit = "7d879bac14d09f2f2a45a0477c1e45fbf52240f5",
    importpath = "github.com/vmware/govmomi",
)

go_repository(
    name = "com_github_weaveworks_mesh",
    commit = "1f158d31de55abf9f97bbaa0a260e2b8023a3785",
    importpath = "github.com/weaveworks/mesh",
)

go_repository(
    name = "com_google_cloud_go",
    commit = "050b16d2314d5fc3d4c9a51e4cd5c7468e77f162",
    importpath = "cloud.google.com/go",
)

go_repository(
    name = "in_gopkg_gcfg_v1",
    commit = "298b7a6a3838f79debfaee8bd3bfb2b8d779e756",
    importpath = "gopkg.in/gcfg.v1",
)

go_repository(
    name = "in_gopkg_inf_v0",
    commit = "3887ee99ecf07df5b447e9b00d9c0b2adaa9f3e4",
    importpath = "gopkg.in/inf.v0",
)

go_repository(
    name = "in_gopkg_natefinch_lumberjack_v2",
    commit = "a96e63847dc3c67d17befa69c303767e2f84e54f",
    importpath = "gopkg.in/natefinch/lumberjack.v2",
)

go_repository(
    name = "in_gopkg_warnings_v0",
    commit = "ec4a0fea49c7b46c2aeb0b51aac55779c607e52b",
    importpath = "gopkg.in/warnings.v0",
)

go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "287cf08546ab5e7e37d55a84f7ed3fd1db036de5",
    importpath = "gopkg.in/yaml.v2",
)

go_repository(
    name = "io_k8s_api",
    commit = "9b9dca205a15b6ce9ef10091f05d60a13fdcf418",
    importpath = "k8s.io/api",
)

go_repository(
    name = "io_k8s_apiextensions_apiserver",
    commit = "9e1143f05c11a1b5f5c2f79434659e9894786cc4",
    importpath = "k8s.io/apiextensions-apiserver",
)

go_repository(
    name = "io_k8s_apimachinery",
    commit = "5134afd2c0c91158afac0d8a28bd2177185a3bcc",
    importpath = "k8s.io/apimachinery",
)

go_repository(
    name = "io_k8s_apiserver",
    commit = "f5fd0005b06697e897cbeb10c97b7cb5f5b85232",
    importpath = "k8s.io/apiserver",
)

go_repository(
    name = "io_k8s_client_go",
    commit = "2ae454230481a7cb5544325e12ad7658ecccd19b",
    importpath = "k8s.io/client-go",
)

go_repository(
    name = "io_k8s_kube_openapi",
    commit = "b16ebc07f5cad97831f961e4b5a9cc1caed33b7e",
    importpath = "k8s.io/kube-openapi",
)

go_repository(
    name = "io_k8s_kubernetes",
    commit = "cce11c6a185279d037023e02ac5249e14daa22bf",
    importpath = "k8s.io/kubernetes",
)

go_repository(
    name = "io_k8s_metrics",
    commit = "270d76282f5b666ea1923500518c1ffa1c5b11ce",
    importpath = "k8s.io/metrics",
)

go_repository(
    name = "io_k8s_utils",
    commit = "66423a0293c555337adc04fe2c59748151291de8",
    importpath = "k8s.io/utils",
)

go_repository(
    name = "ml_vbom_util",
    commit = "256737ac55c46798123f754ab7d2c784e2c71783",
    importpath = "vbom.ml/util",
)

go_repository(
    name = "org_bitbucket_ww_goautoneg",
    commit = "75cd24fc2f2c2a2088577d12123ddee5f54e0675",
    importpath = "bitbucket.org/ww/goautoneg",
)

go_repository(
    name = "org_golang_google_api",
    commit = "03a4a4fe3d2c5fb0d1a116cf20784fffa259a3d4",
    importpath = "google.golang.org/api",
)

go_repository(
    name = "org_golang_google_appengine",
    commit = "150dc57a1b433e64154302bdc40b6bb8aefa313a",
    importpath = "google.golang.org/appengine",
)

go_repository(
    name = "org_golang_google_genproto",
    commit = "a8101f21cf983e773d0c1133ebc5424792003214",
    importpath = "google.golang.org/genproto",
)

go_repository(
    name = "org_golang_google_grpc",
    commit = "e687fa4e6424368ece6e4fe727cea2c806a0fcb4",
    importpath = "google.golang.org/grpc",
)

go_repository(
    name = "org_golang_x_crypto",
    commit = "94eea52f7b742c7cbe0b03b22f0c4c8631ece122",
    importpath = "golang.org/x/crypto",
)

go_repository(
    name = "org_golang_x_net",
    commit = "d866cfc389cec985d6fda2859936a575a55a3ab6",
    importpath = "golang.org/x/net",
)

go_repository(
    name = "org_golang_x_oauth2",
    commit = "462316686f20eb6df426961c1c131bdaa5dfa68e",
    importpath = "golang.org/x/oauth2",
)

go_repository(
    name = "org_golang_x_sys",
    commit = "571f7bbbe08da2a8955aed9d4db316e78630e9a3",
    importpath = "golang.org/x/sys",
)

go_repository(
    name = "org_golang_x_text",
    commit = "d5a9226ed7dd70cade6ccae9d37517fe14dd9fee",
    importpath = "golang.org/x/text",
)
