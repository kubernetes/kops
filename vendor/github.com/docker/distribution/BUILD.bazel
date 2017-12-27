load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "blobs.go",
        "doc.go",
        "errors.go",
        "manifests.go",
        "registry.go",
        "tags.go",
    ],
    importpath = "github.com/docker/distribution",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/docker/distribution/context:go_default_library",
        "//vendor/github.com/docker/distribution/reference:go_default_library",
        "//vendor/github.com/opencontainers/go-digest:go_default_library",
    ],
)
