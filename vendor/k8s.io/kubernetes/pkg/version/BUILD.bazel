load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "base.go",
        "doc.go",
        "version.go",
    ],
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/apimachinery/pkg/version:go_default_library"],
)
