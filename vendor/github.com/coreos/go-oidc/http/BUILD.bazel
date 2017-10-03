load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "client.go",
        "doc.go",
        "http.go",
        "url.go",
    ],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "http_test.go",
        "url_test.go",
    ],
    library = ":go_default_library",
)
