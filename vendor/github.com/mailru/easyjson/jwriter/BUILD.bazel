load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["writer.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/mailru/easyjson/buffer:go_default_library"],
)
