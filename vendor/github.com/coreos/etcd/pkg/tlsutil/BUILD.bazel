load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "tlsutil.go",
    ],
    importmap = "vendor/github.com/coreos/etcd/pkg/tlsutil",
    importpath = "github.com/coreos/etcd/pkg/tlsutil",
    visibility = ["//visibility:public"],
)
