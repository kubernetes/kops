load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "id.go",
        "set.go",
        "slice.go",
        "urls.go",
        "urlsmap.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/types",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "id_test.go",
        "set_test.go",
        "slice_test.go",
        "urls_test.go",
        "urlsmap_test.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/types",
    library = ":go_default_library",
    deps = ["//vendor/github.com/coreos/etcd/pkg/testutil:go_default_library"],
)
