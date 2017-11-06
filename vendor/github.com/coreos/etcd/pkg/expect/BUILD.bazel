load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["expect.go"],
    importpath = "github.com/coreos/etcd/pkg/expect",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/kr/pty:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["expect_test.go"],
    importpath = "github.com/coreos/etcd/pkg/expect",
    library = ":go_default_library",
)
