load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "buffer.go",
        "certs.go",
        "channel.go",
        "cipher.go",
        "client.go",
        "client_auth.go",
        "common.go",
        "connection.go",
        "doc.go",
        "handshake.go",
        "kex.go",
        "keys.go",
        "mac.go",
        "messages.go",
        "mux.go",
        "server.go",
        "session.go",
        "streamlocal.go",
        "tcpip.go",
        "transport.go",
    ],
    importpath = "golang.org/x/crypto/ssh",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/golang.org/x/crypto/curve25519:go_default_library",
        "//vendor/golang.org/x/crypto/ed25519:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "benchmark_test.go",
        "buffer_test.go",
        "certs_test.go",
        "cipher_test.go",
        "client_auth_test.go",
        "client_test.go",
        "handshake_test.go",
        "kex_test.go",
        "keys_test.go",
        "mempipe_test.go",
        "messages_test.go",
        "mux_test.go",
        "session_test.go",
        "tcpip_test.go",
        "testdata_test.go",
        "transport_test.go",
    ],
    importpath = "golang.org/x/crypto/ssh",
    library = ":go_default_library",
    deps = [
        "//vendor/golang.org/x/crypto/ed25519:go_default_library",
        "//vendor/golang.org/x/crypto/ssh/terminal:go_default_library",
        "//vendor/golang.org/x/crypto/ssh/testdata:go_default_library",
    ],
)

go_test(
    name = "go_default_xtest",
    srcs = ["example_test.go"],
    importpath = "golang.org/x/crypto/ssh_test",
    deps = [
        ":go_default_library",
        "//vendor/golang.org/x/crypto/ssh/terminal:go_default_library",
    ],
)
