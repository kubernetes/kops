load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "parse.go",
        "xfs.go",
    ],
    importpath = "github.com/prometheus/procfs/xfs",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["parse_test.go"],
    importpath = "github.com/prometheus/procfs/xfs_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/prometheus/procfs:go_default_library",
    ],
)
