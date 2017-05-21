git_repository(
    name = "io_bazel_rules_go",
    tag = "0.4.4",
    #commit = "d5abc8247362cb27e0b2342486bed3ef6a1f68df",
    remote = "https://github.com/bazelbuild/rules_go.git",
)

load("@io_bazel_rules_go//go:def.bzl", "go_repositories", "go_repository", "new_go_repository")

go_repositories()

#=============================================================================

git_repository(
    name = "org_pubref_rules_protobuf",
    remote = "https://github.com/pubref/rules_protobuf.git",
    tag = "v0.7.1",
)

load("@org_pubref_rules_protobuf//go:rules.bzl", "go_proto_repositories")

go_proto_repositories()

#=============================================================================

# for building docker base images
debs = (
    (
        "deb_busybox",
        "83d809a22d765e52390c0bc352fe30e9d1ac7c82fd509e0d779d8289bfc8a53d",
        "http://ftp.us.debian.org/debian/pool/main/b/busybox/busybox-static_1.22.0-9+deb8u1_amd64.deb",
    ),
    (
        "deb_libc",
        "2d8de90c084a26c266fa8efa91564f99b2373a7949caa9a1db83460918e6e832",
        "http://ftp.us.debian.org/debian/pool/main/g/glibc/libc6_2.19-18+deb8u7_amd64.deb",
    ),
    (
        "deb_ca_certificates",
        "f58d646045855277c87f532ea5c18df319e91d9892437880c9a0169b834f1bd8",
        "http://ftp.us.debian.org/debian/pool/main/c/ca-certificates/ca-certificates_20141019+deb8u1_all.deb",
    ),
)

[http_file(
    name = name,
    sha256 = sha256,
    url = url,
) for name, sha256, url in debs]

#=============================================================================
# client-go

new_go_repository(
    name = "io_k8s_client_go",
    commit = "0389c75147549613c776c45d2de9511339b0c072",
    importpath = "k8s.io/client-go",
)

new_go_repository(
    name = "io_k8s_apiserver",
    commit = "f0eaebe542e55a09c7301fb1da38694f433d9b72",
    importpath = "k8s.io/apiserver",
)

new_go_repository(
    name = "io_k8s_apimachinery",
    commit = "84c15da65eb86243c295d566203d7689cc6ac04b",
    importpath = "k8s.io/apimachinery",
)

#new_go_repository(
#    name = "io_k8s_kubernetes",
##    url = ["https://github.com/kubernetes/kubernetes/archive/8b706690fb90c7e0aade244017e8daa9445cb374.tar.gz"],
#    commit = "8b706690fb90c7e0aade244017e8daa9445cb374",
##    remote = "https://github.com/kubernetes/kubernetes",
#    importpath = "k8s.io/kubernetes",
#)

#http_archive(
#    name = "io_k8s_kubernetes",
#    strip_prefix = "kubernetes-8b706690fb90c7e0aade244017e8daa9445cb374",
#    urls = ["https://github.com/kubernetes/kubernetes/archive/8b706690fb90c7e0aade244017e8daa9445cb374.tar.gz"],
#    #importpath = "k8s.io/kubernetes",
#)

#local_repository(
#    name="io_k8s_kubernetes",
#    path="../kubernetes",
#)

http_archive(
    name = "io_kubernetes_build",
    sha256 = "8d1cff71523565996903076cec6cad8424afa6eb93a342d0d810a55c911e23c7",
    strip_prefix = "repo-infra-61b7247ebf472398bdea148d8f67e3a1849d6de9",
    urls = ["https://github.com/kubernetes/repo-infra/archive/61b7247ebf472398bdea148d8f67e3a1849d6de9.tar.gz"],
)

new_go_repository(
    name = "com_github_PuerkitoBio_purell",
    commit = "8a290539e2e8629dbc4e6bad948158f790ec31f4",
    importpath = "github.com/PuerkitoBio/purell",
)

new_go_repository(
    name = "com_github_PuerkitoBio_urlesc",
    commit = "5bd2802263f21d8788851d5305584c82a5c75d7e",
    importpath = "github.com/PuerkitoBio/urlesc",
)

new_go_repository(
    name = "com_github_coreos_go_oidc",
    commit = "5644a2f50e2d2d5ba0b474bc5bc55fea1925936d",
    importpath = "github.com/coreos/go-oidc",
)

new_go_repository(
    name = "com_github_coreos_pkg",
    commit = "fa29b1d70f0beaddd4c7021607cc3c3be8ce94b8",
    importpath = "github.com/coreos/pkg",
)

new_go_repository(
    name = "com_github_davecgh_go_spew",
    commit = "5215b55f46b2b919f50a1df0eaa5886afe4e3b3d",
    importpath = "github.com/davecgh/go-spew",
)

new_go_repository(
    name = "com_github_docker_distribution",
    commit = "cd27f179f2c10c5d300e6d09025b538c475b0d51",
    importpath = "github.com/docker/distribution",
)

new_go_repository(
    name = "com_github_emicklei_go_restful",
    commit = "09691a3b6378b740595c1002f40c34dd5f218a22",
    importpath = "github.com/emicklei/go-restful",
)

new_go_repository(
    name = "com_github_ghodss_yaml",
    commit = "73d445a93680fa1a78ae23a5839bad48f32ba1ee",
    importpath = "github.com/ghodss/yaml",
)

new_go_repository(
    name = "com_github_go_openapi_jsonpointer",
    commit = "46af16f9f7b149af66e5d1bd010e3574dc06de98",
    importpath = "github.com/go-openapi/jsonpointer",
)

new_go_repository(
    name = "com_github_go_openapi_analysis",
    commit = "b44dc874b601d9e4e2f6e19140e794ba24bead3b",
    importpath = "github.com/go-openapi/analysis",
)

new_go_repository(
    name = "com_github_go_openapi_jsonreference",
    commit = "13c6e3589ad90f49bd3e3bbe2c2cb3d7a4142272",
    importpath = "github.com/go-openapi/jsonreference",
)

new_go_repository(
    name = "com_github_go_openapi_loads",
    commit = "18441dfa706d924a39a030ee2c3b1d8d81917b38",
    importpath = "github.com/go-openapi/loads",
)

new_go_repository(
    name = "com_github_go_openapi_spec",
    commit = "6aced65f8501fe1217321abf0749d354824ba2ff",
    importpath = "github.com/go-openapi/spec",
)

new_go_repository(
    name = "com_github_go_openapi_swag",
    commit = "1d0bd113de87027671077d3c71eb3ac5d7dbba72",
    importpath = "github.com/go-openapi/swag",
)

new_go_repository(
    name = "com_github_gogo_protobuf",
    commit = "c0656edd0d9eab7c66d1eb0c568f9039345796f7",
    importpath = "github.com/gogo/protobuf",
)

new_go_repository(
    name = "com_github_golang_glog",
    commit = "44145f04b68cf362d9c4df2182967c2275eaefed",
    importpath = "github.com/golang/glog",
)

new_go_repository(
    name = "com_github_golang_groupcache",
    commit = "02826c3e79038b59d737d3b1c0a1d937f71a4433",
    importpath = "github.com/golang/groupcache",
)

new_go_repository(
    name = "com_github_golang_protobuf",
    commit = "8616e8ee5e20a1704615e6c8d7afcdac06087a67",
    importpath = "github.com/golang/protobuf",
)

new_go_repository(
    name = "com_github_google_gofuzz",
    commit = "44d81051d367757e1c7c6a5a86423ece9afcf63c",
    importpath = "github.com/google/gofuzz",
)

new_go_repository(
    name = "com_github_howeyc_gopass",
    commit = "3ca23474a7c7203e0a0a070fd33508f6efdb9b3d",
    importpath = "github.com/howeyc/gopass",
)

new_go_repository(
    name = "com_github_imdario_mergo",
    commit = "6633656539c1639d9d78127b7d47c622b5d7b6dc",
    importpath = "github.com/imdario/mergo",
)

new_go_repository(
    name = "com_github_jonboulle_clockwork",
    commit = "72f9bd7c4e0c2a40055ab3d0f09654f730cce982",
    importpath = "github.com/jonboulle/clockwork",
)

new_go_repository(
    name = "com_github_juju_ratelimit",
    commit = "77ed1c8a01217656d2080ad51981f6e99adaa177",
    importpath = "github.com/juju/ratelimit",
)

new_go_repository(
    name = "com_github_mailru_easyjson",
    commit = "d5b7844b561a7bc640052f1b935f7b800330d7e0",
    importpath = "github.com/mailru/easyjson",
)

new_go_repository(
    name = "com_github_pmezard_go_difflib",
    commit = "d8ed2627bdf02c080bf22230dbb337003b7aba2d",
    importpath = "github.com/pmezard/go-difflib",
)

new_go_repository(
    name = "com_github_spf13_pflag",
    commit = "5ccb023bc27df288a957c5e994cd44fd19619465",
    importpath = "github.com/spf13/pflag",
)

new_go_repository(
    name = "com_github_stretchr_testify",
    commit = "e3a8ff8ce36581f87a15341206f205b1da467059",
    importpath = "github.com/stretchr/testify",
)

new_go_repository(
    name = "com_github_ugorji_go",
    commit = "ded73eae5db7e7a0ef6f55aace87a2873c5d2b74",
    importpath = "github.com/ugorji/go",
)

new_go_repository(
    name = "com_google_cloud_go_compute_metadata",
    commit = "3b1ae45394a234c385be014e9a488f2bb6eef821",
    importpath = "cloud.google.com/go/compute/metadata",
)

new_go_repository(
    name = "com_google_cloud_go_internal",
    commit = "3b1ae45394a234c385be014e9a488f2bb6eef821",
    importpath = "cloud.google.com/go/internal",
)

new_go_repository(
    name = "in_gopkg_inf_v0",
    commit = "3887ee99ecf07df5b447e9b00d9c0b2adaa9f3e4",
    importpath = "gopkg.in/inf.v0",
)

new_go_repository(
    name = "in_gopkg_yaml_v2",
    commit = "53feefa2559fb8dfa8d81baad31be332c97d6c77",
    importpath = "gopkg.in/yaml.v2",
)

new_go_repository(
    name = "org_golang_google_appengine",
    commit = "4f7eeb5305a4ba1966344836ba4af9996b7b4e05",
    importpath = "google.golang.org/appengine",
)

new_go_repository(
    name = "org_golang_google_appengine_internal",
    commit = "4f7eeb5305a4ba1966344836ba4af9996b7b4e05",
    importpath = "google.golang.org/appengine/internal",
)

new_go_repository(
    name = "org_golang_google_appengine_internal_app_identity",
    commit = "4f7eeb5305a4ba1966344836ba4af9996b7b4e05",
    importpath = "google.golang.org/appengine/internal/app_identity",
)

new_go_repository(
    name = "org_golang_google_appengine_internal_base",
    commit = "4f7eeb5305a4ba1966344836ba4af9996b7b4e05",
    importpath = "google.golang.org/appengine/internal/base",
)

new_go_repository(
    name = "org_golang_google_appengine_internal_datastore",
    commit = "4f7eeb5305a4ba1966344836ba4af9996b7b4e05",
    importpath = "google.golang.org/appengine/internal/datastore",
)

new_go_repository(
    name = "org_golang_google_appengine_internal_log",
    commit = "4f7eeb5305a4ba1966344836ba4af9996b7b4e05",
    importpath = "google.golang.org/appengine/internal/log",
)

new_go_repository(
    name = "org_golang_google_appengine_internal_modules",
    commit = "4f7eeb5305a4ba1966344836ba4af9996b7b4e05",
    importpath = "google.golang.org/appengine/internal/modules",
)

new_go_repository(
    name = "org_golang_google_appengine_internal_remote_api",
    commit = "4f7eeb5305a4ba1966344836ba4af9996b7b4e05",
    importpath = "google.golang.org/appengine/internal/remote_api",
)

new_go_repository(
    name = "org_golang_x_crypto",
    commit = "d172538b2cfce0c13cee31e647d0367aa8cd2486",
    importpath = "golang.org/x/crypto",
)

new_go_repository(
    name = "org_golang_x_net",
    commit = "e90d6d0afc4c315a0d87a568ae68577cc15149a0",
    importpath = "golang.org/x/net",
)

new_go_repository(
    name = "org_golang_x_oauth2",
    commit = "3c3a985cb79f52a3190fbc056984415ca6763d01",
    importpath = "golang.org/x/oauth2",
)

new_go_repository(
    name = "org_golang_x_sys",
    commit = "8f0908ab3b2457e2e15403d3697c9ef5cb4b57a9",
    importpath = "golang.org/x/sys",
)

new_go_repository(
    name = "org_golang_x_text",
    commit = "2910a502d2bf9e43193af9d68ca516529614eed3",
    importpath = "golang.org/x/text",
)

#=============================================================================
# other deps

new_go_repository(
    name = "com_github_18F_hmacauth",
    commit = "9232a6386b737d7d1e5c1c6e817aa48d5d8ee7cd",
    importpath = "github.com/18F/hmacauth",
)

new_go_repository(
    name = "org_golang_google_api",
    commit = "650535c7d6201e8304c92f38c922a9a3a36c6877",
    importpath = "google.golang.org/api",
)

new_go_repository(
    name = "com_google_cloud_go",
    commit = "dbe4740b523eecbc19b2050f0691772c312aa07b",
    importpath = "cloud.google.com/go",
)

new_go_repository(
    name = "com_github_googleapis_gax_go",
    commit = "8c5154c0fe5bf18cf649634d4c6df50897a32751",
    importpath = "github.com/googleapis/gax-go",
)

new_go_repository(
    name = "com_github_coreos_etcd",
    commit = "cc198e22d3b8fd7ec98304c95e68ee375be54589",
    importpath = "github.com/coreos/etcd",
)

new_go_repository(
    name = "com_github_pborman_uuid",
    commit = "ca53cad383cad2479bbba7f7a1a05797ec1386e4",
    importpath = "github.com/pborman/uuid",
)

new_go_repository(
    name = "com_github_prometheus_client_golang",
    commit = "e51041b3fa41cece0dca035740ba6411905be473",
    importpath = "github.com/prometheus/client_golang",
)

new_go_repository(
    name = "com_github_prometheus_client_model",
    commit = "fa8ad6fec33561be4280a8f0514318c79d7f6cb6",
    importpath = "github.com/prometheus/client_model",
)

new_go_repository(
    name = "com_github_prometheus_common",
    commit = "ffe929a3f4c4faeaa10f2b9535c2b1be3ad15650",
    importpath = "github.com/prometheus/common",
)

new_go_repository(
    name = "com_github_prometheus_procfs",
    commit = "454a56f35412459b5e684fd5ec0f9211b94f002a",
    importpath = "github.com/prometheus/procfs",
)

new_go_repository(
    name = "com_github_beorn7_perks",
    commit = "3ac7bf7a47d159a033b107610db8a1b6575507a4",
    importpath = "github.com/beorn7/perks",
)

new_go_repository(
    name = "org_bitbucket_ww_goautoneg",
    commit = "75cd24fc2f2c2a2088577d12123ddee5f54e0675",
    importpath = "bitbucket.org/ww/goautoneg",
)

new_go_repository(
    name = "com_github_matttproud_golang_protobuf_extensions",
    commit = "c12348ce28de40eed0136aa2b644d0ee0650e56c",
    importpath = "github.com/matttproud/golang_protobuf_extensions",
)

new_go_repository(
    name = "com_github_grpc_ecosystem_grpc_gateway",
    commit = "f52d055dc48aec25854ed7d31862f78913cf17d1",
    importpath = "github.com/grpc-ecosystem/grpc-gateway",
)

new_go_repository(
    name = "com_github_coreos_go_systemd",
    commit = "48702e0da86bd25e76cfef347e2adeb434a0d0a6",
    importpath = "github.com/coreos/go-systemd",
)

new_go_repository(
    name = "com_github_pkg_errors",
    commit = "a22138067af1c4942683050411a841ade67fe1eb",
    importpath = "github.com/pkg/errors",
)

new_go_repository(
    name = "in_gopkg_natefinch_lumberjack_v2",
    commit = "20b71e5b60d756d3d2f80def009790325acc2b23",
    importpath = "gopkg.in/natefinch/lumberjack.v2",
)

new_go_repository(
    name = "com_github_elazarl_go_bindata_assetfs",
    commit = "3dcc96556217539f50599357fb481ac0dc7439b9",
    importpath = "github.com/elazarl/go-bindata-assetfs",
)

new_go_repository(
    name = "com_github_evanphx_json_patch",
    commit = "ba18e35c5c1b36ef6334cad706eb681153d2d379",
    importpath = "github.com/evanphx/json-patch",
)

new_go_repository(
    name = "org_golang_google_grpc",
    commit = "231b4cfea0e79843053a33f5fe90bd4d84b23cd3",
    importpath = "google.golang.org/grpc",
)

new_go_repository(
    name = "com_github_spf13_cobra",
    commit = "f62e98d28ab7ad31d707ba837a966378465c7b57",
    importpath = "github.com/spf13/cobra",
)

new_go_repository(
    name = "com_github_blang_semver",
    commit = "4a1e882c79dcf4ec00d2e29fac74b9c8938d5052",
    importpath = "github.com/blang/semver",
)

new_go_repository(
    name = "com_github_pkg_sftp",
    commit = "a5f8514e29e90a859e93871b1582e5c81f466f82",
    importpath = "github.com/pkg/sftp",
)

new_go_repository(
    name = "com_github_aws_aws_sdk_go",
    commit = "31484500fe77b88dbe197c6348358ed275aed5d7",
    importpath = "github.com/aws/aws-sdk-go",
)

new_go_repository(
    name = "com_github_kr_fs",
    commit = "2788f0dbd16903de03cb8186e5c7d97b69ad387b",
    importpath = "github.com/kr/fs",
)

new_go_repository(
    name = "com_github_go_ini_ini",
    commit = "afbc45e87f3ba324c532d12c71918ef52e0fb194",
    importpath = "github.com/go-ini/ini",
)

new_go_repository(
    name = "com_github_jmespath_go_jmespath",
    commit = "bd40a432e4c76585ef6b72d3fd96fb9b6dc7b68d",
    importpath = "github.com/jmespath/go-jmespath",
)

new_go_repository(
    name = "com_github_sergi_go_diff",
    commit = "feef008d51ad2b3778f85d387ccf91735543008d",
    importpath = "github.com/sergi/go-diff",
)

new_go_repository(
    name = "com_github_inconshreveable_mousetrap",
    commit = "76626ae9c91c4f2a10f34cad8ce83ea42c93bb75",
    importpath = "github.com/inconshreveable/mousetrap",
)
