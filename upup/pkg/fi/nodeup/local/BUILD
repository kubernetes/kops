load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["local_target.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//upup/pkg/fi:go_default_library",
        "@io_k8s_apimachinery//pkg/util/sets:go_default_library",
    ],
)
