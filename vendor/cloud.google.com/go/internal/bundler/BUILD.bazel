load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["bundler.go"],
    importpath = "cloud.google.com/go/internal/bundler",
    visibility = ["//vendor/cloud.google.com/go:__subpackages__"],
)

go_test(
    name = "go_default_test",
    srcs = ["bundler_test.go"],
    importpath = "cloud.google.com/go/internal/bundler",
    library = ":go_default_library",
)
