load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "claims.go",
        "doc.go",
        "ecdsa.go",
        "ecdsa_utils.go",
        "errors.go",
        "hmac.go",
        "map_claims.go",
        "none.go",
        "parser.go",
        "rsa.go",
        "rsa_pss.go",
        "rsa_utils.go",
        "signing_method.go",
        "token.go",
    ],
    importpath = "github.com/dgrijalva/jwt-go",
    visibility = ["//visibility:public"],
)

go_test(
    name = "go_default_xtest",
    srcs = [
        "ecdsa_test.go",
        "example_test.go",
        "hmac_example_test.go",
        "hmac_test.go",
        "http_example_test.go",
        "none_test.go",
        "parser_test.go",
        "rsa_pss_test.go",
        "rsa_test.go",
    ],
    importpath = "github.com/dgrijalva/jwt-go_test",
    deps = [
        ":go_default_library",
        "//vendor/github.com/dgrijalva/jwt-go/request:go_default_library",
        "//vendor/github.com/dgrijalva/jwt-go/test:go_default_library",
    ],
)
