load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "types.go",
    ],
    importmap = "vendor/k8s.io/apimachinery/pkg/version",
    importpath = "k8s.io/apimachinery/pkg/version",
    visibility = ["//visibility:public"],
)
