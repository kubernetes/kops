load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["descriptor.go"],
    importpath = "github.com/golang/protobuf/descriptor",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/descriptor:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["descriptor_test.go"],
    importpath = "github.com/golang/protobuf/descriptor_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/golang/protobuf/proto/testdata:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/descriptor:go_default_library",
    ],
)
