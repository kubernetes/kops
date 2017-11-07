load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "pagewriter.go",
        "readcloser.go",
        "reader.go",
        "util.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/ioutil",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/etcd/pkg/fileutil:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "pagewriter_test.go",
        "readcloser_test.go",
        "reader_test.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/ioutil",
    library = ":go_default_library",
)
