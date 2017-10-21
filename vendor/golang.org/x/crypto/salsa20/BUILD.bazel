load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["salsa20.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/crypto/salsa20/salsa:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["salsa20_test.go"],
    library = ":go_default_library",
)
