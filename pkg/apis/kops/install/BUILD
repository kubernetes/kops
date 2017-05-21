load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["install.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "//pkg/apis/kops/v1alpha1:go_default_library",
        "//pkg/apis/kops/v1alpha2:go_default_library",
        "@io_k8s_apimachinery//pkg/apimachinery/announced:go_default_library",
        "@io_k8s_apimachinery//pkg/apimachinery/registered:go_default_library",
        "@io_k8s_apimachinery//pkg/runtime:go_default_library",
        "@io_k8s_apimachinery//pkg/util/sets:go_default_library",
    ],
)
