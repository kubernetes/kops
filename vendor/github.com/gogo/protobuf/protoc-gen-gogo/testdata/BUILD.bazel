load("@io_bazel_rules_go//go:def.bzl", "go_test")

go_test(
    name = "go_default_test",
    srcs = [
        "extension_test.go",
        "main_test.go",
    ],
    importpath = "github.com/gogo/protobuf/protoc-gen-gogo/testdata",
    deps = [
        "//vendor/github.com/gogo/protobuf/protoc-gen-gogo/testdata/multi:go_default_library",
        "//vendor/github.com/gogo/protobuf/protoc-gen-gogo/testdata/my_test:go_default_library",
    ],
)
