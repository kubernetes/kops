load("@io_bazel_rules_go//go:def.bzl", "go_library", "go_test")

go_library(
    name = "go_default_library",
    srcs = [
        "flags.go",
        "util.go",
        "viper.go",
    ],
    importpath = "github.com/spf13/viper",
    visibility = ["//visibility:public"],
    deps = [
        "//vendor/github.com/fsnotify/fsnotify:go_default_library",
        "//vendor/github.com/hashicorp/hcl:go_default_library",
        "//vendor/github.com/magiconair/properties:go_default_library",
        "//vendor/github.com/mitchellh/mapstructure:go_default_library",
        "//vendor/github.com/pelletier/go-toml:go_default_library",
        "//vendor/github.com/spf13/afero:go_default_library",
        "//vendor/github.com/spf13/cast:go_default_library",
        "//vendor/github.com/spf13/jwalterweatherman:go_default_library",
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/gopkg.in/yaml.v2:go_default_library",
    ],
)

go_test(
    name = "go_default_test",
    srcs = [
        "flags_test.go",
        "viper_test.go",
    ],
    importpath = "github.com/spf13/viper",
    library = ":go_default_library",
    deps = [
        "//vendor/github.com/spf13/pflag:go_default_library",
        "//vendor/github.com/stretchr/testify/assert:go_default_library",
    ],
)
