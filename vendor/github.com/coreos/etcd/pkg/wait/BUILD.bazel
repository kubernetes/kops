load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "wait.go",
        "wait_time.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/wait",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "wait_test.go",
        "wait_time_test.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/wait",
    library = ":go_default_library",
)
