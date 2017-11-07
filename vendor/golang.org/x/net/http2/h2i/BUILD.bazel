load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = ["h2i.go"],
    importpath = "golang.org/x/net/http2/h2i",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/golang.org/x/crypto/ssh/terminal:go_default_library",
        "//vendor/golang.org/x/net/http2:go_default_library",
        "//vendor/golang.org/x/net/http2/hpack:go_default_library",
    ],
)

go_binary(
    name = "h2i",
    importpath = "golang.org/x/net/http2/h2i",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
