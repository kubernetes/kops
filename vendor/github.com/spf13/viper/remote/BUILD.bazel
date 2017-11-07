load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["remote.go"],
    importpath = "github.com/spf13/viper/remote",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/spf13/viper:go_default_library",
        "//vendor/github.com/xordataexchange/crypt/config:go_default_library",
    ],
)
