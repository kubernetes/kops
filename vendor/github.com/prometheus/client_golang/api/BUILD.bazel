load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["client.go"],
    importpath = "github.com/prometheus/client_golang/api",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["client_test.go"],
    importpath = "github.com/prometheus/client_golang/api",
    library = ":go_default_library",
)
