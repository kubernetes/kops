load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "glog.go",
        "glog_file.go",
    ],
    importpath = "github.com/golang/glog",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["glog_test.go"],
    importpath = "github.com/golang/glog",
    library = ":go_default_library",
)
