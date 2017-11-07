load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "log.go",
        "log_unstable.go",
        "logger.go",
        "node.go",
        "progress.go",
        "raft.go",
        "rawnode.go",
        "read_only.go",
        "status.go",
        "storage.go",
        "util.go",
    ],
    importpath = "github.com/coreos/etcd/raft",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "diff_test.go",
        "example_test.go",
        "log_test.go",
        "log_unstable_test.go",
        "node_bench_test.go",
        "node_test.go",
        "progress_test.go",
        "raft_flow_control_test.go",
        "raft_paper_test.go",
        "raft_snap_test.go",
        "raft_test.go",
        "rawnode_test.go",
        "storage_test.go",
        "util_test.go",
    ],
    importpath = "github.com/coreos/etcd/raft",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/testutil:go_default_library",
        "//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
