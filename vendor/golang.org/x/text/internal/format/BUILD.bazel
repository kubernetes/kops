load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["format.go"],
    visibility = ["//vendor/golang.org/x/text:__subpackages__"],
    deps = ["//vendor/golang.org/x/text/language:go_default_library"],
)
