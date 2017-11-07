load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "httpapi.go",
        "kvstore.go",
        "listener.go",
        "main.go",
        "raft.go",
    ],
    importpath = "github.com/coreos/etcd/contrib/raftexample",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/coreos/etcd/etcdserver/stats:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/fileutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/github.com/coreos/etcd/raft:go_default_library",
        "//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library",
        "//vendor/github.com/coreos/etcd/rafthttp:go_default_library",
        "//vendor/github.com/coreos/etcd/snap:go_default_library",
        "//vendor/github.com/coreos/etcd/wal:go_default_library",
        "//vendor/github.com/coreos/etcd/wal/walpb:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_binary(
    name = "raftexample",
    importpath = "github.com/coreos/etcd/contrib/raftexample",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "kvstore_test.go",
        "raftexample_test.go",
    ],
    importpath = "github.com/coreos/etcd/contrib/raftexample",
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library"],
)
