load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["testutils.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/docker/docker/api/types/events:go_default_library",
        "//vendor/github.com/docker/docker/api/types/time:go_default_library",
    ],
)
