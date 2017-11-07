load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["mixbench.go"],
    importpath = "github.com/gogo/protobuf/test/mixbench",
    visibility = ["//visibility:private"],
)

go_binary(
    name = "mixbench",
    importpath = "github.com/gogo/protobuf/test/mixbench",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
