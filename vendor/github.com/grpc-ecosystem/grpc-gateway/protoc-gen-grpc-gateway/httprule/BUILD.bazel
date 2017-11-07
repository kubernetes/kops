load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "compile.go",
        "parse.go",
        "types.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/utilities:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "compile_test.go",
        "parse_test.go",
        "types_test.go",
    ],
    importpath = "github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway/httprule",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/utilities:go_default_library",
    ],
)
