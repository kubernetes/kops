load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["syncer.go"],
    importpath = "github.com/coreos/etcd/clientv3/mirror",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/clientv3:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
