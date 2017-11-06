load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "cloud.google.com/go/examples/bigtable/helloworld",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/cloud.google.com/go/bigtable:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_binary(
    name = "helloworld",
    importpath = "cloud.google.com/go/examples/bigtable/helloworld",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
