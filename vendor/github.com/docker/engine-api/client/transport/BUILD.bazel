load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "client.go",
        "transport.go",
    ],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/docker/go-connections/sockets:go_default_library"],
)
