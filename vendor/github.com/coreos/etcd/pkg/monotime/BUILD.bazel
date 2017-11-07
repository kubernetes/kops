load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "issue15006.s",
        "monotime.go",
        "nanotime.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/monotime",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["nanotime_test.go"],
    importpath = "github.com/coreos/etcd/pkg/monotime",
    library = ":go_default_library",
)
