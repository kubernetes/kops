load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "link_grpc.go",
        "main.go",
    ],
    importpath = "github.com/golang/protobuf/protoc-gen-go",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/generator:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/grpc:go_default_library",
    ],
)

go_binary(
    name = "protoc-gen-go",
    importpath = "github.com/golang/protobuf/protoc-gen-go",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
