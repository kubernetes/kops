load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "bytes.go",
        "more_test_objects.pb.go",
        "test_objects.pb.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gogo/protobuf/proto:go_default_library",
        "//vendor/github.com/gogo/protobuf/types:go_default_library",
    ],
)

filegroup(
    name = "go_default_library_protos",
    srcs = [
        "more_test_objects.proto",
        "test_objects.proto",
    ],
    visibility = ["//visibility:public"],
)
