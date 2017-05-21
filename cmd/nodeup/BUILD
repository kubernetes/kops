load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    visibility = ["//visibility:private"],
    deps = [
        "//:go_default_library",
        "//nodeup/pkg/bootstrap:go_default_library",
        "//upup/models:go_default_library",
        "//upup/pkg/fi/nodeup:go_default_library",
        "@com_github_golang_glog//:go_default_library",
    ],
)

go_binary(
    name = "nodeup",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
