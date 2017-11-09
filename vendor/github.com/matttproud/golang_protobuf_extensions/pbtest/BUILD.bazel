load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "quick.go",
    ],
    importpath = "github.com/matttproud/golang_protobuf_extensions/pbtest",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = ["example_test.go"],
    importpath = "github.com/matttproud/golang_protobuf_extensions/pbtest",
    library = ":go_default_library",
    deps = ["//vendor/github.com/golang/protobuf/proto:go_default_library"],
)
