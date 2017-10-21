load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "config.go",
        "seccomp_unsupported.go",
    ],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/opencontainers/runc/libcontainer/configs:go_default_library"],
)
