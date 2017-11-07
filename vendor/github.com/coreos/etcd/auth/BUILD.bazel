load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "range_perm_cache.go",
        "simple_token.go",
        "store.go",
    ],
    importpath = "github.com/coreos/etcd/auth",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/auth/authpb:go_default_library",
        "//vendor/github.com/coreos/etcd/etcdserver/etcdserverpb:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
        "//vendor/golang.org/x/crypto/bcrypt:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/google.golang.org/grpc/metadata:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "range_perm_cache_test.go",
        "store_test.go",
    ],
    importpath = "github.com/coreos/etcd/auth",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/etcdserver/etcdserverpb:go_default_library",
        "//vendor/github.com/coreos/etcd/mvcc/backend:go_default_library",
        "//vendor/golang.org/x/crypto/bcrypt:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
