load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["timeseries.go"],
    visibility = ["//vendor/golang.org/x/net:__subpackages__"],
)

go_test(
    name = "go_default_test",
    srcs = ["timeseries_test.go"],
    library = ":go_default_library",
)
