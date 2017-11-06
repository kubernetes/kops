load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["struct.pb.go"],
    importpath = "github.com/golang/protobuf/ptypes/struct",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

proto_library(
    name = "structpb_proto",
    srcs = ["struct.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "structpb_go_proto",
    importpath = "github.com/golang/protobuf/ptypes/struct",
    proto = ":structpb_proto",
    visibility = ["//visibility:public"],
)
