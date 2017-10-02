load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["httplex.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/net/idna:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["httplex_test.go"],
    library = ":go_default_library",
)
