load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["serverreflection.go"],
    importpath = "google.golang.org/grpc/reflection",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/descriptor:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
        "//vendor/google.golang.org/grpc/reflection/grpc_reflection_v1alpha:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["serverreflection_test.go"],
    importpath = "google.golang.org/grpc/reflection",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/descriptor:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/reflection/grpc_reflection_v1alpha:go_default_library",
        "//vendor/google.golang.org/grpc/reflection/grpc_testing:go_default_library",
    ],
)
