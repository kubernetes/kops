load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "catmsg.go",
        "codec.go",
        "varint.go",
    ],
    importpath = "golang.org/x/text/internal/catmsg",
    visibility = ["//vendor/golang.org/x/text:__subpackages__"],
    deps = ["//vendor/golang.org/x/text/language:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "catmsg_test.go",
        "varint_test.go",
    ],
    importpath = "golang.org/x/text/internal/catmsg",
    library = ":go_default_library",
    deps = ["//vendor/golang.org/x/text/language:go_default_library"],
)
