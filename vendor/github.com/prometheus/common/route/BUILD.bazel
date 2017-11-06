load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["route.go"],
    importpath = "github.com/prometheus/common/route",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/julienschmidt/httprouter:go_default_library",
        "//vendor/golang.org/x/net/context:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["route_test.go"],
    importpath = "github.com/prometheus/common/route",
    library = ":go_default_library",
)
