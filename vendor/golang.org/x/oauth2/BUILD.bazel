load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "oauth2.go",
        "token.go",
        "transport.go",
    ],
    importpath = "golang.org/x/oauth2",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/net/context:go_default_library",
        "//vendor/golang.org/x/oauth2/internal:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "oauth2_test.go",
        "token_test.go",
        "transport_test.go",
    ],
    importpath = "golang.org/x/oauth2",
    library = ":go_default_library",
    deps = ["//vendor/golang.org/x/net/context:go_default_library"],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "golang.org/x/oauth2_test",
    deps = [":go_default_library"],
)
