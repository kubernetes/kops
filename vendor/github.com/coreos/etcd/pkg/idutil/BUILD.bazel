load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["id.go"],
    importpath = "github.com/coreos/etcd/pkg/idutil",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["id_test.go"],
    importpath = "github.com/coreos/etcd/pkg/idutil",
    library = ":go_default_library",
)
