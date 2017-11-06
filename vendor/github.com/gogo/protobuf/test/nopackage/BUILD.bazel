load("@io_bazel_rules_go//proto:def.bzl", "go_proto_library")
load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["nopackage.pb.go"],
    importpath = "github.com/gogo/protobuf/test/nopackage",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/gogo/protobuf/proto:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["nopackage_test.go"],
    importpath = "github.com/gogo/protobuf/test/nopackage",
    library = ":go_default_library",
)

proto_library(
    name = "nopackage_proto",
    srcs = ["nopackage.proto"],
    visibility = ["//visibility:public"],
)

go_proto_library(
    name = "nopackage_go_proto",
    importpath = "github.com/gogo/protobuf/test/nopackage",
    proto = ":nopackage_proto",
    visibility = ["//visibility:public"],
)
