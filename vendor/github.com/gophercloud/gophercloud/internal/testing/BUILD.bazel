load("@io_bazel_rules_go//go:def.bzl", "go_test")

go_test(
    name = "go_default_test",
    srcs = ["util_test.go"],
    importpath = "github.com/gophercloud/gophercloud/internal/testing",
    deps = ["//vendor/github.com/gophercloud/gophercloud/internal:go_default_library"],
)
