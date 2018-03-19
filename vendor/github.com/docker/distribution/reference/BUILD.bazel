load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "helpers.go",
        "normalize.go",
        "reference.go",
        "regexp.go",
    ],
    importpath = "github.com/docker/distribution/reference",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/docker/distribution/digestset:go_default_library",
        "//vendor/github.com/opencontainers/go-digest:go_default_library",
    ],
)
