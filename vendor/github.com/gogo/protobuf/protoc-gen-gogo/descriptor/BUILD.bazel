load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "descriptor.go",
        "descriptor.pb.go",
        "descriptor_gostring.gen.go",
        "helper.go",
    ],
    importpath = "github.com/gogo/protobuf/protoc-gen-gogo/descriptor",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/gogo/protobuf/proto:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["descriptor_test.go"],
    importpath = "github.com/gogo/protobuf/protoc-gen-gogo/descriptor_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/gogo/protobuf/proto/testdata:go_default_library",
    ],
)
