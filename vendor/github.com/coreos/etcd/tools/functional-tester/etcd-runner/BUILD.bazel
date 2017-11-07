load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "help.go",
        "main.go",
    ],
    importpath = "github.com/coreos/etcd/tools/functional-tester/etcd-runner",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/coreos/etcd/tools/functional-tester/etcd-runner/command:go_default_library",
        "//vendor/github.com/coreos/etcd/version:go_default_library",
        "//vendor/github.com/spf13/cobra:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
    ],
)

go_binary(
    name = "etcd-runner",
    importpath = "github.com/coreos/etcd/tools/functional-tester/etcd-runner",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
