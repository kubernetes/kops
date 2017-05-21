load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "clientset.go",
        "cluster.go",
        "federation.go",
        "instancegroup.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops:go_default_library",
        "@io_k8s_apimachinery//pkg/apis/meta/v1:go_default_library",
    ],
)
