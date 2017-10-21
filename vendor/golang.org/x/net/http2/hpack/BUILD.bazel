load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "encode.go",
        "hpack.go",
        "huffman.go",
        "tables.go",
    ],
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "encode_test.go",
        "hpack_test.go",
        "tables_test.go",
    ],
    library = ":go_default_library",
)
