load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["consistenthash.go"],
    importpath = "github.com/golang/groupcache/consistenthash",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["consistenthash_test.go"],
    importpath = "github.com/golang/groupcache/consistenthash",
    library = ":go_default_library",
)
