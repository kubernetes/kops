load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["types.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/k8s.io/kubernetes/pkg/api/unversioned:go_default_library",
        "//vendor/k8s.io/kubernetes/pkg/api/v1:go_default_library",
    ],
)
