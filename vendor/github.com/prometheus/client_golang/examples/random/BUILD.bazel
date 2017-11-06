load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/prometheus/client_golang/examples/random",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus/promhttp:go_default_library",
    ],
)

go_binary(
    name = "random",
    importpath = "github.com/prometheus/client_golang/examples/random",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
