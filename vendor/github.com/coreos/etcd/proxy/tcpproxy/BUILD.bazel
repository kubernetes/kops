load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "userspace.go",
    ],
    importpath = "github.com/coreos/etcd/proxy/tcpproxy",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/pkg/capnslog:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["userspace_test.go"],
    importpath = "github.com/coreos/etcd/proxy/tcpproxy",
    library = ":go_default_library",
)
