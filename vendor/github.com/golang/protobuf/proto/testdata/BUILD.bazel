load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["test.pb.go"],
    importpath = "github.com/golang/protobuf/proto/testdata",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["golden_test.go"],
    importpath = "github.com/golang/protobuf/proto/testdata",
    library = ":go_default_library",
)

proto_library(
    name = "testdata_proto",
    srcs = ["test.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "testdata_go_proto",
    importpath = "github.com/golang/protobuf/proto/testdata",
    proto = ":testdata_proto",
    visibility = ["//visibility:public"],
)
