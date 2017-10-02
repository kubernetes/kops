load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["middleware.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/aws/aws-sdk-go/service/cloudfront/sign:go_default_library",
        "//vendor/github.com/docker/distribution/context:go_default_library",
        "//vendor/github.com/docker/distribution/registry/storage/driver:go_default_library",
        "//vendor/github.com/docker/distribution/registry/storage/driver/middleware:go_default_library",
    ],
)
