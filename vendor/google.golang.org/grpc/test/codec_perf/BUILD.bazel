load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["perf.pb.go"],
    importpath = "google.golang.org/grpc/test/codec_perf",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

proto_library(
    name = "codec_perf_proto",
    srcs = ["perf.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "codec_perf_go_proto",
    importpath = "google.golang.org/grpc/test/codec_perf",
    proto = ":codec_perf_proto",
    visibility = ["//visibility:public"],
)
