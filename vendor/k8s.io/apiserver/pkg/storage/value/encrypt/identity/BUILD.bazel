load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["identity.go"],
    importpath = "k8s.io/apiserver/pkg/storage/value/encrypt/identity",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/apiserver/pkg/storage/value:go_default_library"],
)
