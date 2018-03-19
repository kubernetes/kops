load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["httplex.go"],
    importpath = "golang.org/x/net/lex/httplex",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/net/idna:go_default_library"],
)
