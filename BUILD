load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_prefix")
load("@io_bazel_rules_go//go:def.bzl", "go_prefix")

go_prefix("k8s.io/kops")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "version.go",
    ],
    visibility = ["//visibility:public"],
)
