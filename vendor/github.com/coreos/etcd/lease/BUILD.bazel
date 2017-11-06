load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "lessor.go",
    ],
    importpath = "github.com/coreos/etcd/lease",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/lease/leasepb:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/monotime:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["lessor_test.go"],
    importpath = "github.com/coreos/etcd/lease",
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library"],
)
