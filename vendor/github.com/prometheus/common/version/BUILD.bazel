load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["info.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/prometheus/client_golang/prometheus:go_default_library"],
)
