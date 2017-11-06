load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["hello_dolly.go"],
    importpath = "github.com/coreos/pkg/capnslog/example",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/coreos/pkg/capnslog:go_default_library"],
)

go_binary(
    name = "example",
    importpath = "github.com/coreos/pkg/capnslog/example",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
