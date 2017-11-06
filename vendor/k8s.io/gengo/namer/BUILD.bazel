load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "import_tracker.go",
        "namer.go",
        "order.go",
        "plural_namer.go",
    ],
    importpath = "k8s.io/gengo/namer",
    visibility = ["//visibility:public"],
    deps = ["//vendor/k8s.io/gengo/types:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "namer_test.go",
        "plural_namer_test.go",
    ],
    importpath = "k8s.io/gengo/namer",
    library = ":go_default_library",
    deps = ["//vendor/k8s.io/gengo/types:go_default_library"],
)
