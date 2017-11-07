load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "wait_recorder.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/mock/mockwait",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/testutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/wait:go_default_library",
    ],
)
