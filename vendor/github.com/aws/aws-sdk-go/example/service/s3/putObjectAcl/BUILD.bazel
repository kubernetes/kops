load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["putObjectAcl.go"],
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/aws/aws-sdk-go/aws:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/aws/session:go_default_library",
        "//vendor/github.com/aws/aws-sdk-go/service/s3:go_default_library",
    ],
)

go_binary(
    name = "putObjectAcl",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
