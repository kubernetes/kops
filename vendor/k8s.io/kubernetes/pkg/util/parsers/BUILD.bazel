load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["parsers.go"],
    importpath = "k8s.io/kubernetes/pkg/util/parsers",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/docker/distribution/reference:go_default_library"],
)
