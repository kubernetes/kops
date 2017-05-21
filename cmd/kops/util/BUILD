load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["factory.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//pkg/apis/kops/install:go_default_library",
        "//pkg/client/simple:go_default_library",
        "//pkg/client/simple/vfsclientset:go_default_library",
        "//util/pkg/vfs:go_default_library",
        "@io_k8s_apimachinery//pkg/util/validation/field:go_default_library",
    ],
)
