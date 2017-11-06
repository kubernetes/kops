load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "errors.go",
        "member.go",
    ],
    importpath = "github.com/coreos/etcd/etcdserver/api/v2http/httptypes",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "errors_test.go",
        "member_test.go",
    ],
    importpath = "github.com/coreos/etcd/etcdserver/api/v2http/httptypes",
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/etcd/pkg/types:go_default_library"],
)
