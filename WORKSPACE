workspace(name = "at_mgit_prometheus_mgit_exporter")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "rules_mgit",
    sha256 = "5a7cfa2cec436d42806c31dfac64cdda9c9c12cf926bc1672f0b556bcec86fbf",
    strip_prefix = "rules_mgit-f3c11b1556c02d00e8a4b2aca730edf0474a7368",
    type = "zip",
    urls = ["https://github.com/mgit-at/rules_mgit/archive/f3c11b1556c02d00e8a4b2aca730edf0474a7368.zip"],
)

load("@rules_mgit//:deps.bzl", "rules_mgit_dependencies")

rules_mgit_dependencies()

load("@rules_mgit//:setup.bzl", "rules_mgit_setup")

rules_mgit_setup()

load("//:go_deps.bzl", "go_repositories")

# gazelle:repository_macro go_deps.bzl%go_repositories
go_repositories()
