load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["start.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/v1alpha2:go_default_library",
        "//pkg/apiserver:go_default_library",
        "//pkg/cmdutil:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@com_github_spf13_cobra//:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime/schema:go_default_library",
        "@io_k8s_apimachinery//pkg/util/wait:go_default_library",
        "@io_k8s_apiserver//pkg/registry/generic:go_default_library",
        "@io_k8s_apiserver//pkg/registry/generic/registry:go_default_library",
        "@io_k8s_apiserver//pkg/server:go_default_library",
        "@io_k8s_apiserver//pkg/server/options:go_default_library",
        "@io_k8s_apiserver//pkg/storage/storagebackend:go_default_library",
    ],
)
