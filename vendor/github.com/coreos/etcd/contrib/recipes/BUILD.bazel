load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "barrier.go",
        "client.go",
        "double_barrier.go",
        "key.go",
        "priority_queue.go",
        "queue.go",
        "rwmutex.go",
        "watch.go",
    ],
    importpath = "github.com/coreos/etcd/contrib/recipes",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/clientv3:go_default_library",
        "//vendor/github.com/coreos/etcd/clientv3/concurrency:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/mvccpb:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
