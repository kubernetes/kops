load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["cloud_init_target.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//upup/pkg/fi:go_default_library",
        "//upup/pkg/fi/utils:go_default_library",
        "@com_github_golang_glog//:go_default_library",
        "@io_k8s_apimachinery//pkg/util/sets:go_default_library",
    ],
)
