load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["main.go"],
    importpath = "github.com/gogo/protobuf/protoc-gen-gogo",
    visibility = ["//visibility:private"],
    deps = ["//vendor/github.com/gogo/protobuf/vanity/command:go_default_library"],
)

go_binary(
    name = "protoc-gen-gogo",
    importpath = "github.com/gogo/protobuf/protoc-gen-gogo",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
