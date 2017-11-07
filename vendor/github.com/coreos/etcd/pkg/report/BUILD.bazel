load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "report.go",
        "timeseries.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/report",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "report_test.go",
        "timeseries_test.go",
    ],
    importpath = "github.com/coreos/etcd/pkg/report",
    library = ":go_default_library",
)
