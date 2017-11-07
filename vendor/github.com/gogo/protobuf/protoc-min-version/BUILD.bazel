load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["minversion.go"],
    importpath = "github.com/gogo/protobuf/protoc-min-version",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/gogo/protobuf/version:go_default_library"],
)

go_binary(
    name = "protoc-min-version",
    importpath = "github.com/gogo/protobuf/protoc-min-version",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
