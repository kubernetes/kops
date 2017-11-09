load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "lease.go",
        "mvcc.go",
        "mvcc-put.go",
        "put.go",
        "range.go",
        "root.go",
        "stm.go",
        "util.go",
        "watch.go",
        "watch_get.go",
        "watch_latency.go",
    ],
    importpath = "github.com/coreos/etcd/tools/benchmark/cmd",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/clientv3:go_default_library",
        "//vendor/github.com/coreos/etcd/clientv3/concurrency:go_default_library",
        "//vendor/github.com/coreos/etcd/lease:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/report:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/transport:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/time/rate:go_default_library",
        "//vendor/gopkg.in/cheggaaa/pb.v1:go_default_library",
    ],
)
