load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/coreos/etcd",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/coreos/etcd/etcdmain:go_default_library"],
)

go_binary(
    name = "etcd",
    importpath = "github.com/coreos/etcd",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
