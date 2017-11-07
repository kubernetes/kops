load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "config.go",
        "doc.go",
        "etcd.go",
        "serve.go",
        "util.go",
    ],
    importpath = "github.com/coreos/etcd/embed",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/cockroachdb/cmux:go_default_library",
        "//vendor/github.com/coreos/etcd/discovery:go_default_library",
        "//vendor/github.com/coreos/etcd/etcdserver:go_default_library",
        "//vendor/github.com/coreos/etcd/etcdserver/api/v2http:go_default_library",
        "//vendor/github.com/coreos/etcd/etcdserver/api/v3rpc:go_default_library",
        "//vendor/github.com/coreos/etcd/etcdserver/etcdserverpb:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/cors:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/netutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/runtime:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/transport:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/github.com/coreos/etcd/rafthttp:go_default_library",
        "//vendor/github.com/coreos/etcd/wal:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
        "//vendor/github.com/ghodss/yaml:go_default_library",
        "//vendor/github.com/grpc-ecosystem/grpc-gateway/runtime:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/credentials:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["config_test.go"],
    importpath = "github.com/coreos/etcd/embed",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/transport:go_default_library",
        "//vendor/github.com/ghodss/yaml:go_default_library",
    ],
)
