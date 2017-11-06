load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["wrappers.pb.go"],
    importpath = "github.com/golang/protobuf/ptypes/wrappers",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

proto_library(
    name = "wrappers_proto",
    srcs = ["wrappers.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "wrappers_go_proto",
    importpath = "github.com/golang/protobuf/ptypes/wrappers",
    proto = ":wrappers_proto",
    visibility = ["//visibility:public"],
)
