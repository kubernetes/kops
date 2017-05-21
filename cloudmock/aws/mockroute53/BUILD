load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "api.go",
        "records.go",
        "unimplemented.go",
        "zones.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/request:go_default_library",
        "@com_github_aws_aws_sdk_go//service/route53:go_default_library",
        "@com_github_aws_aws_sdk_go//service/route53/route53iface:go_default_library",
        "@com_github_golang_glog//:go_default_library",
    ],
)
