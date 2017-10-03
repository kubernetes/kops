load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "main.go",
        "template.go",
    ],
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/docker/docker/pkg/aaparser:go_default_library"],
)

go_binary(
    name = "apparmor",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
