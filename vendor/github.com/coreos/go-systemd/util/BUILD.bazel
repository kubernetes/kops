load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "util.go",
        "util_cgo.go",
    ],
    cgo = True,
    importpath = "github.com/coreos/go-systemd/util",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/coreos/pkg/dlopen:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["util_test.go"],
    importpath = "github.com/coreos/go-systemd/util",
    library = ":go_default_library",
)
