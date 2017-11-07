load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "error.go",
        "md.go",
    ],
    importpath = "github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["error_test.go"],
    importpath = "github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes",
    library = ":go_default_library",
    deps = [
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/codes:go_default_library",
    ],
)
