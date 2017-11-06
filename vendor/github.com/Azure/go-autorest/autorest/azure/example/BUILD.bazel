load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/Azure/go-autorest/autorest/azure/example",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/Azure/go-autorest/autorest:go_default_library",
        "//vendor/github.com/Azure/go-autorest/autorest/adal:go_default_library",
        "//vendor/github.com/Azure/go-autorest/autorest/azure:go_default_library",
        "//vendor/golang.org/x/crypto/pkcs12:go_default_library",
    ],
)

go_binary(
    name = "example",
    importpath = "github.com/Azure/go-autorest/autorest/azure/example",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
