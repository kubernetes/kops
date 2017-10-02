load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["utils_unix.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/docker/docker/pkg/stringutils:go_default_library",
        "//vendor/github.com/syndtr/gocapability/capability:go_default_library",
    ],
)
