load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "main.go",
    ],
    importpath = "github.com/coreos/etcd/tools/etcd-dump-logs",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/coreos/etcd/etcdserver/etcdserverpb:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/pbutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library",
        "//vendor/github.com/coreos/etcd/snap:go_default_library",
        "//vendor/github.com/coreos/etcd/wal:go_default_library",
        "//vendor/github.com/coreos/etcd/wal/walpb:go_default_library",
    ],
)

go_binary(
    name = "etcd-dump-logs",
    importpath = "github.com/coreos/etcd/tools/etcd-dump-logs",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
