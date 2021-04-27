workspace(name = "at_mgit_prometheus_mgit_exporter")

load("@bazel_tools//tools/build_defs/repo:http.bzl", "http_archive")

http_archive(
    name = "rules_mgit",
    sha256 = "2f1111adb2faf67ad48170f9ab2a728fb1b7e60973ad1a2b561f19b434eb6faa",
    strip_prefix = "rules_mgit-4a7ed03a8a094d7fbd0daee9ef5d7c68fc00c764",
    type = "zip",
    urls = ["https://github.com/mgit-at/rules_mgit/archive/4a7ed03a8a094d7fbd0daee9ef5d7c68fc00c764.zip"],
)

load("@rules_mgit//:deps.bzl", "rules_mgit_dependencies")

rules_mgit_dependencies()

load("@rules_mgit//:setup.bzl", "rules_mgit_setup")

rules_mgit_setup()

load("//:go_deps.bzl", "go_repositories")

# gazelle:repository_macro go_deps.bzl%go_repositories
go_repositories()
