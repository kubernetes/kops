load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "auth.go",
        "commit.go",
    ],
    importpath = "github.com/Azure/go-autorest/autorest/utils",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/Azure/go-autorest/autorest:go_default_library",
        "//vendor/github.com/Azure/go-autorest/autorest/adal:go_default_library",
        "//vendor/github.com/Azure/go-autorest/autorest/azure:go_default_library",
    ],
)
