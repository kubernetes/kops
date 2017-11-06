load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["adal.go"],
    importpath = "github.com/Azure/go-autorest/autorest/adal/cmd",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/Azure/go-autorest/autorest/adal:go_default_library",
        "//vendor/golang.org/x/crypto/pkcs12:go_default_library",
    ],
)

go_binary(
    name = "cmd",
    importpath = "github.com/Azure/go-autorest/autorest/adal/cmd",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
