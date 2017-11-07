load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/prometheus/client_golang/examples/simple",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/prometheus/client_golang/prometheus/promhttp:go_default_library"],
)

go_binary(
    name = "simple",
    importpath = "github.com/prometheus/client_golang/examples/simple",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
