load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "index.go",
        "key_index.go",
        "kv.go",
        "kvstore.go",
        "kvstore_compaction.go",
        "metrics.go",
        "revision.go",
        "util.go",
        "watchable_store.go",
        "watcher.go",
        "watcher_group.go",
    ],
    importpath = "github.com/coreos/etcd/mvcc",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/lease:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/mvccpb:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/adt:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/schedule:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
        "//vendor/github.com/google/btree:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "index_test.go",
        "key_index_test.go",
        "kv_test.go",
        "kvstore_bench_test.go",
        "kvstore_compaction_test.go",
        "kvstore_test.go",
        "revision_test.go",
        "watchable_store_bench_test.go",
        "watchable_store_test.go",
        "watcher_bench_test.go",
        "watcher_test.go",
    ],
    importpath = "github.com/coreos/etcd/mvcc",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/lease:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/mvccpb:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/schedule:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/testutil:go_default_library",
        "//vendor/github.com/google/btree:go_default_library",
    ],
)
