load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "doc.go",
        "extractor.go",
        "oauth2.go",
        "request.go",
    ],
    importpath = "github.com/dgrijalva/jwt-go/request",
    visibility = ["//visibility:public"],
    deps = ["//vendor/github.com/dgrijalva/jwt-go:go_default_library"],
)

go_test(
    name = "go_default_test",
    srcs = [
        "extractor_example_test.go",
        "extractor_test.go",
        "request_test.go",
    ],
    importpath = "github.com/dgrijalva/jwt-go/request",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/dgrijalva/jwt-go:go_default_library",
        "//vendor/github.com/dgrijalva/jwt-go/test:go_default_library",
    ],
)
