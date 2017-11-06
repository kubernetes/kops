load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "proxy.go",
        "url.go",
    ],
    importpath = "github.com/coreos/pkg/netutil",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/pkg/capnslog:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["url_test.go"],
    importpath = "github.com/coreos/pkg/netutil",
    library = ":go_default_library",
)
