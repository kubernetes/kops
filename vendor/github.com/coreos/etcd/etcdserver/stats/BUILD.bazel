load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "leader.go",
        "queue.go",
        "server.go",
        "stats.go",
    ],
    importpath = "github.com/coreos/etcd/etcdserver/stats",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/raft:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
    ],
)
