load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "agent.go",
        "doc.go",
        "main.go",
        "rpc.go",
    ],
    importpath = "github.com/coreos/etcd/tools/functional-tester/etcd-agent",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/fileutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/netutil:go_default_library",
        "//vendor/github.com/coreos/etcd/tools/functional-tester/etcd-agent/client:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
    ],
)

go_binary(
    name = "etcd-agent",
    importpath = "github.com/coreos/etcd/tools/functional-tester/etcd-agent",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "agent_test.go",
        "rpc_test.go",
    ],
    importpath = "github.com/coreos/etcd/tools/functional-tester/etcd-agent",
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/etcd/tools/functional-tester/etcd-agent/client:go_default_library"],
)
