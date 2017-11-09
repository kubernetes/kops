load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["verflag.go"],
    importpath = "k8s.io/kubernetes/pkg/version/verflag",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/version:go_default_library",
    ],
)
