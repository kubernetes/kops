load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["pointer.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/go-openapi/swag:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["pointer_test.go"],
    library = ":go_default_library",
    deps = ["//vendor/github.com/stretchr/testify/assert:go_default_library"],
)
