load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "cloud.google.com/go/examples/bigquery/query",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/cloud.google.com/go/bigquery:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_binary(
    name = "query",
    importpath = "cloud.google.com/go/examples/bigquery/query",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
