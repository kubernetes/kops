load("@io_bazel_rules_go//go:def.bzl", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "parser.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "parser_unix.go",
        ],
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "parser_unix.go",
        ],
        "@io_bazel_rules_go//go/platform:windows_amd64": [
            "parser_windows.go",
        ],
        "//conditions:default": [],
    }),
    importpath = "github.com/mailru/easyjson/parser",
    visibility = ["//visibility:public"],
)
