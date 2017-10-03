load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "customtype.go",
        "subproto.pb.go",
    ],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/gogo/protobuf/gogoproto:go_default_library",
        "//vendor/github.com/gogo/protobuf/proto:go_default_library",
    ],
)

filegroup(
    name = "go_default_library_protos",
    srcs = ["subproto.proto"],
    visibility = ["//visibility:public"],
)
