load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["volumes.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//protokube/pkg/protokube:go_default_library",
        "//upup/pkg/fi/utils:go_default_library",
        "@com_github_golang_glog//:go_default_library",
    ],
)
