load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "clone.go",
        "decode.go",
        "decode_gogo.go",
        "duration.go",
        "duration_gogo.go",
        "encode.go",
        "encode_gogo.go",
        "equal.go",
        "extensions.go",
        "extensions_gogo.go",
        "lib.go",
        "lib_gogo.go",
        "message_set.go",
        "pointer_unsafe.go",
        "pointer_unsafe_gogo.go",
        "properties.go",
        "properties_gogo.go",
        "skip_gogo.go",
        "text.go",
        "text_gogo.go",
        "text_parser.go",
        "timestamp.go",
        "timestamp_gogo.go",
    ],
    importpath = "github.com/gogo/protobuf/proto",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "message_set_test.go",
        "size2_test.go",
    ],
    importpath = "github.com/gogo/protobuf/proto",
    library = ":go_default_library",
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "all_test.go",
        "any_test.go",
        "clone_test.go",
        "decode_test.go",
        "encode_test.go",
        "equal_test.go",
        "extensions_test.go",
        "map_test.go",
        "proto3_test.go",
        "size_test.go",
        "text_parser_test.go",
        "text_test.go",
    ],
    importpath = "github.com/gogo/protobuf/proto_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/gogo/protobuf/proto/proto3_proto:go_default_library",
        "//vendor/github.com/gogo/protobuf/proto/testdata:go_default_library",
        "//vendor/github.com/gogo/protobuf/types:go_default_library",
    ],
)
