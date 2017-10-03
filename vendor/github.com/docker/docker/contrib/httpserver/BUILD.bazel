load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["server.go"],
    visibility = ["//visibility:private"],
)

go_binary(
    name = "httpserver",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
