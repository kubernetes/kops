load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "helpers.go",
        "mocks.go",
    ],
    importpath = "github.com/Azure/go-autorest/autorest/mocks",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "helpers_test.go",
        "mocks_test.go",
    ],
    importpath = "github.com/Azure/go-autorest/autorest/mocks",
    library = ":go_default_library",
)
