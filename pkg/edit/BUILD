load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["edit.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/diff:go_default_library",
        "//upup/pkg/fi/utils:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime:go_default_library",
    ],
)
