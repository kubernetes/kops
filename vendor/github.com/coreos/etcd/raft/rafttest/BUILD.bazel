load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "network.go",
        "node.go",
    ],
    importpath = "github.com/coreos/etcd/raft/rafttest",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/raft:go_default_library",
        "//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "network_test.go",
        "node_bench_test.go",
        "node_test.go",
    ],
    importpath = "github.com/coreos/etcd/raft/rafttest",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/raft:go_default_library",
        "//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
