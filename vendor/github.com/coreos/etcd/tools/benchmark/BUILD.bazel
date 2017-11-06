load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "main.go",
    ],
    importpath = "github.com/coreos/etcd/tools/benchmark",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/coreos/etcd/tools/benchmark/cmd:go_default_library"],
)

go_binary(
    name = "benchmark",
    importpath = "github.com/coreos/etcd/tools/benchmark",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
