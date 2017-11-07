load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["merge_logger.go"],
    importpath = "github.com/coreos/etcd/pkg/logutil",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/pkg/capnslog:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["merge_logger_test.go"],
    importpath = "github.com/coreos/etcd/pkg/logutil",
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/pkg/capnslog:go_default_library"],
)
