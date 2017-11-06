load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "auth_role.go",
        "auth_user.go",
        "cancelreq.go",
        "client.go",
        "cluster_error.go",
        "curl.go",
        "discover.go",
        "doc.go",
        "keys.generated.go",
        "keys.go",
        "members.go",
        "srv.go",
        "util.go",
    ],
    importpath = "github.com/coreos/etcd/client",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/pathutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/github.com/ugorji/go/codec:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "client_test.go",
        "fake_transport_test.go",
        "keys_bench_test.go",
        "keys_test.go",
        "members_test.go",
        "srv_test.go",
    ],
    importpath = "github.com/coreos/etcd/client",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/testutil:go_default_library",
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)
