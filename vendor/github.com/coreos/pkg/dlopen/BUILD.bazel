load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "dlopen.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:linux_amd64": [
            "dlopen_example.go",
        ],
        "//conditions:default": [],
    }),
    cgo = True,
    clinkopts = ["-ldl"],
    importpath = "github.com/coreos/pkg/dlopen",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_test",
    srcs = ["dlopen_test.go"],
    importpath = "github.com/coreos/pkg/dlopen",
    library = ":go_default_library",
)
