load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "registry.go",
        "services.go",
        "types.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/descriptor:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/generator:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/plugin:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "registry_test.go",
        "services_test.go",
        "types_test.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/descriptor",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/descriptor:go_default_library",
        "//vendor/github.com/golang/protobuf/protoc-gen-go/plugin:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule:go_default_library",
    ],
)
