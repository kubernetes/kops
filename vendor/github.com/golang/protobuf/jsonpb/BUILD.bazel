load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["jsonpb.go"],
    importpath = "github.com/golang/protobuf/jsonpb",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["jsonpb_test.go"],
    importpath = "github.com/golang/protobuf/jsonpb",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/jsonpb/jsonpb_test_proto:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/proto/proto3_proto:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/any:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/duration:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/struct:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/timestamp:go_default_library",
        "//vendor/github.com/golang/protobuf/ptypes/wrappers:go_default_library",
    ],
)
