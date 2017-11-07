load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "authorization.go",
        "autorest.go",
        "client.go",
        "error.go",
        "preparer.go",
        "responder.go",
        "sender.go",
        "utility.go",
        "version.go",
    ],
    importpath = "github.com/Azure/go-autorest/autorest",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/Azure/go-autorest/autorest/adal:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "authorization_test.go",
        "autorest_test.go",
        "client_test.go",
        "error_test.go",
        "preparer_test.go",
        "responder_test.go",
        "sender_test.go",
        "utility_test.go",
    ],
    importpath = "github.com/Azure/go-autorest/autorest",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/Azure/go-autorest/autorest/adal:go_default_library",
        "//vendor/github.com/Azure/go-autorest/autorest/mocks:go_default_library",
    ],
)
