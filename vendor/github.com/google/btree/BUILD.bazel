load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["btree.go"],
    importpath = "github.com/google/btree",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["btree_test.go"],
    importpath = "github.com/google/btree",
    library = ":go_default_library",
)
