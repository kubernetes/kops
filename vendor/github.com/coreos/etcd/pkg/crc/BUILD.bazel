load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["crc.go"],
    importpath = "github.com/coreos/etcd/pkg/crc",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["crc_test.go"],
    importpath = "github.com/coreos/etcd/pkg/crc",
    library = ":go_default_library",
)
