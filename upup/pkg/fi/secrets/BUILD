load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["vfs_secretstore.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//upup/pkg/fi:go_default_library",
        "//util/pkg/vfs:go_default_library",
        "@com_github_golang_glog//:go_default_library",
    ],
)
