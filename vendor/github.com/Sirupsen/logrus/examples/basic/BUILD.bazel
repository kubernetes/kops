load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["basic.go"],
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/Sirupsen/logrus:go_default_library"],
)

go_binary(
    name = "basic",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
