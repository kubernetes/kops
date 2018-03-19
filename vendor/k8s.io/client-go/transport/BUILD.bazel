load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "cache.go",
        "config.go",
        "round_trippers.go",
        "transport.go",
    ],
    importpath = "k8s.io/client-go/transport",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/golang/glog:go_default_library",
        "//vendor/github.com/gregjones/httpcache:go_default_library",
        "//vendor/github.com/gregjones/httpcache/diskcache:go_default_library",
        "//vendor/github.com/peterbourgon/diskv:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/net:go_default_library",
    ],
)
