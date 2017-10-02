load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interface.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/aws/aws-sdk-go/aws/request:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/service/route53domains:go_default_library",
    ],
)
