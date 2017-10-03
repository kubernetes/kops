load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["getter.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/docker/docker/pkg/plugins:go_default_library"],
)
