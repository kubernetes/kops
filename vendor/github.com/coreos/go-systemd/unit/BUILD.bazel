load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "deserialize.go",
        "escape.go",
        "option.go",
        "serialize.go",
    ],
    importpath = "github.com/coreos/go-systemd/unit",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "deserialize_test.go",
        "end_to_end_test.go",
        "escape_test.go",
        "option_test.go",
        "serialize_test.go",
    ],
    importpath = "github.com/coreos/go-systemd/unit",
    library = ":go_default_library",
)
