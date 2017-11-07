load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["cast5.go"],
    importpath = "golang.org/x/crypto/cast5",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["cast5_test.go"],
    importpath = "golang.org/x/crypto/cast5",
    library = ":go_default_library",
)
