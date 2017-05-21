load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    visibility = ["//visibility:private"],
    deps = [
        "//pkg/apiserver/cmd/server:go_default_library",
        "//pkg/cmdutil:go_default_library",
        "@io_k8s_apiserver//pkg/util/logs:go_default_library",
    ],
)

go_binary(
    name = "kops-server",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
