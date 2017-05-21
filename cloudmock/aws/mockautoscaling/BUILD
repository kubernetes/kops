load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "api.go",
        "group.go",
        "unimplemented.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/request:go_default_library",
        "@com_github_aws_aws_sdk_go//service/autoscaling:go_default_library",
        "@com_github_aws_aws_sdk_go//service/autoscaling/autoscalingiface:go_default_library",
    ],
)
