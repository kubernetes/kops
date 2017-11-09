load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "schedule.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/schedule",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/net/context:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["schedule_test.go"],
    importpath = "github.com/coreos/etcd/pkg/schedule",
    library = ":go_default_library",
    deps = ["//vendor/golang.org/x/net/context:go_default_library"],
)
