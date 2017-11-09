load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "http.go",
    ],
    importpath = "github.com/coreos/etcd/lease/leasehttp",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/etcdserver/etcdserverpb:go_default_library",
        "//vendor/github.com/coreos/etcd/lease:go_default_library",
        "//vendor/github.com/coreos/etcd/lease/leasepb:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/httputil:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["http_test.go"],
    importpath = "github.com/coreos/etcd/lease/leasehttp",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/lease:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
