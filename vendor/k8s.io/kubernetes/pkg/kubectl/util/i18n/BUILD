package(default_visibility = ["//visibility:public"])

load(
    "@io_bazel_rules_go//go:def.bzl",
    "go_library",
)

go_library(
    name = "go_default_library",
    srcs = ["i18n.go"],
    importmap = "k8s.io/kops/vendor/k8s.io/kubernetes/pkg/kubectl/util/i18n",
    importpath = "k8s.io/kubernetes/pkg/kubectl/util/i18n",
    deps = [
        "//vendor/github.com/chai2010/gettext-go/gettext:go_default_library",
        "//vendor/k8s.io/klog:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/kubectl/generated:go_default_library",
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
    srcs = [":package-srcs"],
    tags = ["automanaged"],
)
