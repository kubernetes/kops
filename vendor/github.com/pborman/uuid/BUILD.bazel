load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "dce.go",
        "doc.go",
        "hash.go",
        "json.go",
        "node.go",
        "time.go",
        "util.go",
        "uuid.go",
        "version1.go",
        "version4.go",
    ],
    importpath = "github.com/pborman/uuid",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "json_test.go",
        "seq_test.go",
        "uuid_test.go",
    ],
    importpath = "github.com/pborman/uuid",
    library = ":go_default_library",
)
