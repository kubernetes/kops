load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "backend.go",
        "doc.go",
        "main.go",
        "utils.go",
    ],
    importpath = "github.com/coreos/etcd/tools/etcd-dump-db",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/boltdb/bolt:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
    ],
)

go_binary(
    name = "etcd-dump-db",
    importpath = "github.com/coreos/etcd/tools/etcd-dump-db",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
