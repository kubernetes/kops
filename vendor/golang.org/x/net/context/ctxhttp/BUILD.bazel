load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "ctxhttp.go",
        "ctxhttp_pre17.go",
    ],
    importpath = "golang.org/x/net/context/ctxhttp",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/net/context:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "ctxhttp_17_test.go",
        "ctxhttp_pre17_test.go",
        "ctxhttp_test.go",
    ],
    importpath = "golang.org/x/net/context/ctxhttp",
    library = ":go_default_library",
    deps = ["//vendor/golang.org/x/net/context:go_default_library"],
)
