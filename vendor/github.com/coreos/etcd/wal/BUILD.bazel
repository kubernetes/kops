load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "decoder.go",
        "doc.go",
        "encoder.go",
        "file_pipeline.go",
        "metrics.go",
        "repair.go",
        "util.go",
        "wal.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "wal_unix.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "wal_unix.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "wal_windows.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "github.com/coreos/etcd/wal",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/crc:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/fileutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/ioutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/pbutil:go_default_library",
        "//vendor/github.com/coreos/etcd/raft:go_default_library",
        "//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library",
        "//vendor/github.com/coreos/etcd/wal/walpb:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "record_test.go",
        "repair_test.go",
        "wal_bench_test.go",
        "wal_test.go",
    ],
    importpath = "github.com/coreos/etcd/wal",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/fileutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/pbutil:go_default_library",
        "//vendor/github.com/coreos/etcd/raft/raftpb:go_default_library",
        "//vendor/github.com/coreos/etcd/wal/walpb:go_default_library",
    ],
)
