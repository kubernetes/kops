load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "compactor.go",
        "doc.go",
    ],
    importpath = "github.com/coreos/etcd/compactor",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/etcdserver/etcdserverpb:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
        "//vendor/github.com/jonboulle/clockwork:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["compactor_test.go"],
    importpath = "github.com/coreos/etcd/compactor",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/etcdserver/etcdserverpb:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/testutil:go_default_library",
        "//vendor/github.com/jonboulle/clockwork:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
