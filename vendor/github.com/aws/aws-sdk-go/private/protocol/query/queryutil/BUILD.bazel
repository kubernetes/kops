load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["queryutil.go"],
    importpath = "github.com/aws/aws-sdk-go/private/protocol/query/queryutil",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/aws/aws-sdk-go/private/protocol:go_default_library"],
)
