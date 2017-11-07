load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["super-simple-store.go"],
    importpath = "github.com/peterbourgon/diskv/examples/super-simple-store",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/peterbourgon/diskv:go_default_library"],
)

go_binary(
    name = "super-simple-store",
    importpath = "github.com/peterbourgon/diskv/examples/super-simple-store",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
