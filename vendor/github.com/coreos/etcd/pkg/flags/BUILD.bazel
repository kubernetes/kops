load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "flag.go",
        "strings.go",
        "urls.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/flags",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/coreos/etcd/pkg/types:go_default_library",
        "//vendor/github.com/coreos/pkg/capnslog:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "flag_test.go",
        "strings_test.go",
        "urls_test.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/flags",
    library = ":go_default_library",
)
