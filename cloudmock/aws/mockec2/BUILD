load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "address.go",
        "api.go",
        "convenience.go",
        "images.go",
        "keypairs.go",
        "routetable.go",
        "securitygroups.go",
        "subnets.go",
        "tags.go",
        "unimplemented.go",
        "volumes.go",
        "vpcs.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "@com_github_aws_aws_sdk_go//aws:go_default_library",
        "@com_github_aws_aws_sdk_go//aws/request:go_default_library",
        "@com_github_aws_aws_sdk_go//service/ec2:go_default_library",
        "@com_github_aws_aws_sdk_go//service/ec2/ec2iface:go_default_library",
        "@com_github_golang_glog//:go_default_library",
    ],
)
