load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["apiserver.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/install:go_default_library",
        "//pkg/apis/kops/v1alpha2:go_default_library",
        "//pkg/apiserver/registry/cluster:go_default_library",
        "//pkg/cmdutil:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime/schema:go_default_library",
        "@io_k8s_apimachinery//pkg/version:go_default_library",
        "@io_k8s_apiserver//pkg/registry/generic:go_default_library",
        "@io_k8s_apiserver//pkg/registry/rest:go_default_library",
        "@io_k8s_apiserver//pkg/server:go_default_library",
    ],
)
