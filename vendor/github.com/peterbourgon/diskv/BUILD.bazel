load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "compression.go",
        "diskv.go",
        "index.go",
    ],
    importpath = "github.com/peterbourgon/diskv",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/google/btree:go_default_library"],
)
