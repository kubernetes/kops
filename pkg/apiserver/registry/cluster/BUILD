load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "etcd.go",
        "strategy.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "@io_k8s_apimachinery//pkg/fields:go_default_library",
        "@io_k8s_apimachinery//pkg/labels:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime:go_default_library",
        "@io_k8s_apimachinery//pkg/util/validation/field:go_default_library",
        "@io_k8s_apiserver//pkg/endpoints/request:go_default_library",
        "@io_k8s_apiserver//pkg/registry/generic:go_default_library",
        "@io_k8s_apiserver//pkg/registry/generic/registry:go_default_library",
        "@io_k8s_apiserver//pkg/storage:go_default_library",
        "@io_k8s_apiserver//pkg/storage/names:go_default_library",
    ],
)
