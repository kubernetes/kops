load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["diskcache.go"],
    importpath = "github.com/gregjones/httpcache/diskcache",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/peterbourgon/diskv:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["diskcache_test.go"],
    importpath = "github.com/gregjones/httpcache/diskcache",
    library = ":go_default_library",
)
