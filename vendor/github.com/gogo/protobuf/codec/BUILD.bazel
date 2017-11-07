load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["codec.go"],
    importpath = "github.com/gogo/protobuf/codec",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/gogo/protobuf/proto:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["codec_test.go"],
    importpath = "github.com/gogo/protobuf/codec",
    library = ":go_default_library",
    deps = ["//vendor/github.com/gogo/protobuf/test:go_default_library"],
)
