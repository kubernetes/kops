load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "data.go",
        "nothing.go",
        "omitempty.go",
        "snake.go",
    ],
    importpath = "github.com/mailru/easyjson/tests",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/mailru/easyjson:go_default_library",
        "//vendor/github.com/mailru/easyjson/opt:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "basic_test.go",
        "required_test.go",
    ],
    importpath = "github.com/mailru/easyjson/tests",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/mailru/easyjson:go_default_library",
        "//vendor/github.com/mailru/easyjson/jwriter:go_default_library",
    ],
)
