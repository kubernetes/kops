load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["interfaces.go"],
    importpath = "k8s.io/apiserver/pkg/authentication/authenticator",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/apiserver/pkg/authentication/user:go_default_library"],
)
