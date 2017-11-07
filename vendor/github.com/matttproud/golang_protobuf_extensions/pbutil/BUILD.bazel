load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "decode.go",
        "doc.go",
        "encode.go",
    ],
    importpath = "github.com/matttproud/golang_protobuf_extensions/pbutil",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "all_test.go",
        "fixtures_test.go",
    ],
    importpath = "github.com/matttproud/golang_protobuf_extensions/pbutil",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/golang/protobuf/proto:go_default_library",
        "//vendor/github.com/golang/protobuf/proto/testdata:go_default_library",
        "//vendor/github.com/matttproud/golang_protobuf_extensions/pbtest:go_default_library",
    ],
)
