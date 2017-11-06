load("@io_bazel_rules_go//go:def.bzl", "go_binary", "go_library")

go_library(
    name = "go_default_library",
    srcs = [
        "chars.go",
        "col.go",
        "colcmp.go",
        "gen.go",
    ] + select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "darwin.go",
        ],
        "//conditions:default": [],
    }),
    cgo = True,
    clinkopts = select({
        "@io_bazel_rules_go//go/platform:darwin_amd64": [
            "-framework CoreFoundation",
        ],
        "//conditions:default": [],
    }),
    importpath = "golang.org/x/text/collate/tools/colcmp",
    visibility = ["//visibility:private"],
    deps = [
        "//vendor/golang.org/x/text/collate:go_default_library",
        "//vendor/golang.org/x/text/language:go_default_library",
        "//vendor/golang.org/x/text/unicode/norm:go_default_library",
    ],
)

go_binary(
    name = "colcmp",
    importpath = "golang.org/x/text/collate/tools/colcmp",
    library = ":go_default_library",
    visibility = ["//visibility:public"],
)
