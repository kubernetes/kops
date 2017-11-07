load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "event.go",
        "event_history.go",
        "event_queue.go",
        "metrics.go",
        "node.go",
        "node_extern.go",
        "stats.go",
        "store.go",
        "ttl_key_heap.go",
        "watcher.go",
        "watcher_hub.go",
    ],
    importpath = "github.com/coreos/etcd/store",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/error:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/github.com/jonboulle/clockwork:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "event_test.go",
        "heap_test.go",
        "node_extern_test.go",
        "node_test.go",
        "stats_test.go",
        "store_bench_test.go",
        "store_test.go",
        "watcher_hub_test.go",
        "watcher_test.go",
    ],
    importpath = "github.com/coreos/etcd/store",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/error:go_default_library",
        "//vendor/github.com/jonboulle/clockwork:go_default_library",
        "//vendor/github.com/stretchr/testify/assert:go_default_library",
    ],
)
