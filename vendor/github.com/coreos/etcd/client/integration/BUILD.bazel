load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["doc.go"],
    importpath = "github.com/coreos/etcd/client/integration",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "client_test.go",
        "main_test.go",
    ],
    importpath = "github.com/coreos/etcd/client/integration",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/client:go_default_library",
        "//vendor/github.com/coreos/etcd/integration:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/testutil:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
