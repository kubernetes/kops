load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["file.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/pkg/sftp:go_default_library"],
)
