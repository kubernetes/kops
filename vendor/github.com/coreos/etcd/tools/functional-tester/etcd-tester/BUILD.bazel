load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "checks.go",
        "cluster.go",
        "doc.go",
        "failpoint.go",
        "failure.go",
        "failure_agent.go",
        "http.go",
        "key_stresser.go",
        "lease_stresser.go",
        "main.go",
        "member.go",
        "metrics.go",
        "status.go",
        "stresser.go",
        "tester.go",
        "util.go",
        "v2_stresser.go",
    ],
    importpath = "github.com/coreos/etcd/tools/functional-tester/etcd-tester",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/github.com/coreos/etcd/client:go_default_library",
        "//vendor/github.com/coreos/etcd/clientv3:go_default_library",
        "//vendor/github.com/coreos/etcd/etcdserver:go_default_library",
        "//vendor/github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes:go_default_library",
        "//vendor/github.com/coreos/etcd/etcdserver/etcdserverpb:go_default_library",
        "//vendor/github.com/coreos/etcd/tools/functional-tester/etcd-agent/client:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
        "//vendor/github.com/prometheus/client_golang/prometheus:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/time/rate:go_default_library",
        "//vendor/google.golang.org/grpc:go_default_library",
        "//vendor/google.golang.org/grpc/grpclog:go_default_library",
        "//vendor/google.golang.org/grpc/transport:go_default_library",
    ],
)

go_binary(
    name = "etcd-tester",
    importpath = "github.com/coreos/etcd/tools/functional-tester/etcd-tester",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
