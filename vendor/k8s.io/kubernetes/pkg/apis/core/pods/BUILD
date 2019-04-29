load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["helpers.go"],
    importmap = "k8s.io/kops/vendor/k8s.io/kubernetes/pkg/apis/core/pods",
    importpath = "k8s.io/kubernetes/pkg/apis/core/pods",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/kubernetes/pkg/fieldpath:go_default_library"],
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
    visibility = ["//visibility:public"],
)
