load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "httplog.go",
    ],
    importpath = "k8s.io/apiserver/pkg/server/httplog",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/glog:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["httplog_test.go"],
    importpath = "k8s.io/apiserver/pkg/server/httplog",
    library = ":go_default_library",
)
