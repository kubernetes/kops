load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "crdfinder.go",
        "factory.go",
        "factory_client_access.go",
        "helpers.go",
        "kubectl_match_version.go",
        "printing.go",
    ],
    importmap = "k8s.io/kops/vendor/k8s.io/kubernetes/pkg/kubectl/cmd/util",
    importpath = "k8s.io/kubernetes/pkg/kubectl/cmd/util",
    visibility = ["//build/visible_to:pkg_kubectl_cmd_util_CONSUMERS"],
    deps = [
        "//vendor/github.com/evanphx/json-patch:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/k8s.io/api/core/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/api/meta:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/apis/meta/v1/unstructured:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/runtime/schema:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/errors:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/sets:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/yaml:go_default_library",
        "//vendor/k8s.io/cli-runtime/pkg/genericclioptions:go_default_library",
        "//vendor/k8s.io/cli-runtime/pkg/resource:go_default_library",
        "//vendor/k8s.io/client-go/discovery:go_default_library",
        "//vendor/k8s.io/client-go/dynamic:go_default_library",
        "//vendor/k8s.io/client-go/kubernetes:go_default_library",
        "//vendor/k8s.io/client-go/rest:go_default_library",
        "//vendor/k8s.io/client-go/scale:go_default_library",
        "//vendor/k8s.io/client-go/tools/clientcmd:go_default_library",
        "//vendor/k8s.io/klog:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi/validation:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/kubectl/scheme:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/kubectl/util/templates:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/kubectl/validation:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/kubectl/version:go_default_library",
        "//vendor/k8s.io/utils/exec:go_default_library",
    ],
)

filegroup(
    name = "package-srcs",
    srcs = glob(["**"]),
    tags = ["automanaged"],
    visibility = ["//visibility:private"],
)

filegroup(
    name = "all-srcs",
    srcs = [
        ":package-srcs",
        "//pkg/kubectl/cmd/util/editor:all-srcs",
        "//pkg/kubectl/cmd/util/openapi:all-srcs",
        "//pkg/kubectl/cmd/util/sanity:all-srcs",
    ],
    tags = ["automanaged"],
    visibility = ["//build/visible_to:pkg_kubectl_cmd_util_CONSUMERS"],
)
