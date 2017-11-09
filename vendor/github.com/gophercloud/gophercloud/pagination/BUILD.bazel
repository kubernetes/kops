load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "http.go",
        "linked.go",
        "marker.go",
        "pager.go",
        "pkg.go",
        "single.go",
    ],
    importpath = "github.com/gophercloud/gophercloud/pagination",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/gophercloud/gophercloud:go_default_library"],
)
