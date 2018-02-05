load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "const.go",
        "doc.go",
        "doctype.go",
        "entity.go",
        "escape.go",
        "foreign.go",
        "node.go",
        "parse.go",
        "render.go",
        "token.go",
    ],
    importpath = "golang.org/x/net/html",
    visibility = ["//visibility:public"],
    deps = ["//vendor/golang.org/x/net/html/atom:go_default_library"],
)
