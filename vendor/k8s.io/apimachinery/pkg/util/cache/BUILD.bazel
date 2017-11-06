load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "cache.go",
        "lruexpirecache.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/util/cache",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/hashicorp/golang-lru:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "cache_test.go",
        "lruexpirecache_test.go",
    ],
    importpath = "k8s.io/apimachinery/pkg/util/cache",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/groupcache/lru:go_default_library",
        "//vendor/k8s.io/apimachinery/pkg/util/clock:go_default_library",
    ],
)
