load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "keepalive_listener.go",
        "limit_listen.go",
        "listener.go",
        "timeout_conn.go",
        "timeout_dialer.go",
        "timeout_listener.go",
        "timeout_transport.go",
        "tls.go",
        "transport.go",
        "unix_listener.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/transport",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/fileutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/tlsutil:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "keepalive_listener_test.go",
        "listener_test.go",
        "timeout_dialer_test.go",
        "timeout_listener_test.go",
        "timeout_transport_test.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/transport",
    library = ":go_default_library",
)
