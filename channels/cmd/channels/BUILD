load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    visibility = ["//visibility:private"],
    deps = ["//channels/pkg/cmd:go_default_library"],
)

go_binary(
    name = "channels",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
