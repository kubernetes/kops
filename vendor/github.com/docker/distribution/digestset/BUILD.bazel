load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["set.go"],
    importpath = "github.com/docker/distribution/digestset",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/opencontainers/go-digest:go_default_library"],
)
