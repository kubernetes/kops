load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["healthz.go"],
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/google/cadvisor/http/mux:go_default_library"],
)
