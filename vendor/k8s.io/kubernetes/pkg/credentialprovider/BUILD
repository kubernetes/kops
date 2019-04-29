package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = [
        "config.go",
        "doc.go",
        "keyring.go",
        "plugins.go",
        "provider.go",
    ],
    importmap = "k8s.io/kops/vendor/k8s.io/kubernetes/pkg/credentialprovider",
    importpath = "k8s.io/kubernetes/pkg/credentialprovider",
    deps = [
        "//vendor/k8s.io/apimachinery/pkg/util/sets:go_default_library",
        "//vendor/k8s.io/klog:go_default_library",
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
        "//pkg/credentialprovider/aws:all-srcs",
        "//pkg/credentialprovider/azure:all-srcs",
        "//pkg/credentialprovider/gcp:all-srcs",
        "//pkg/credentialprovider/rancher:all-srcs",
        "//pkg/credentialprovider/secrets:all-srcs",
    ],
    tags = ["automanaged"],
)
