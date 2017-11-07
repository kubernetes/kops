load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = ["clientcredentials.go"],
    importpath = "golang.org/x/oauth2/clientcredentials",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/oauth2:go_default_library",
        "//vendor/golang.org/x/oauth2/internal:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = ["clientcredentials_test.go"],
    importpath = "golang.org/x/oauth2/clientcredentials",
    library = ":go_default_library",
)
