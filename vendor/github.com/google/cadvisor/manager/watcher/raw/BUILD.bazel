load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["raw.go"],
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/google/cadvisor/container/common:go_default_library",
        "//vendor/github.com/google/cadvisor/container/libcontainer:go_default_library",
        "//vendor/github.com/google/cadvisor/manager/watcher:go_default_library",
        "//vendor/golang.org/x/exp/inotify:go_default_library",
    ],
)
