load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["option.go"],
    importpath = "cloud.google.com/go/bigtable/internal/option",
    visibility = ["//vendor/cloud.google.com/go/bigtable:__subpackages__"],
    deps = [
        "//vendor/google.golang.org/api/option:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
    ],
)
