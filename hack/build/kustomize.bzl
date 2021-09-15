# Copyright 2020 The Cockroach Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     https://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# The kustomization rule generates rules to kustomize the current target.
#
# Targets:
#   <name> - generates the base to a single file at <out>
#   <name>.preview - generate the base and print to stdout
#
# The main goal of this rule is to make it easy to use kustomize to build manifests that can be used as
# k8s_deploy/k8s_object(s) dependencies.
#
# Example:
#
#     load("@k8s_deploy//:defaults.bzl", "k8s_deploy")
#     load("//hack/build:kustomize.bzl", "kustomization")
#
#     kustomization(
#         name = "manifest",
#         srcs = [":all-srcs"],
#     )
#
#     k8s_deploy(
#         name = "crd",
#         template = ":manifest",
#         visibility = ["//visibility:public"],
#     )
def kustomization(
    name,
    srcs,
    out = "manifest.yaml",
    tools = ["//hack/boilerplate:all-srcs", "//hack/bin:kustomize"],
    visibility = ["//visibility:public"]):
  rule_srcs = "{}.srcs".format(name)
  rule_main = "{}.main".format(name)

  native.filegroup(name = rule_srcs, srcs = srcs, visibility = ["//visibility:public"])
  native.filegroup(name = rule_main, srcs = ["kustomization.yaml"], visibility = ["//visibility:private"])

  build_cmd = " ".join([
      "$(location //hack/bin:kustomize)",
      "build",
      "--load-restrictor=LoadRestrictionsNone", # allow symlinks because bazel
      "$$(dirname $(location :{}))".format(rule_main)])

  # bazel build <name> to generate manifest.yaml
  native.genrule(
      name = name,
      exec_tools = tools,
      srcs = [":{}".format(rule_srcs), ":{}".format(rule_main)],
      outs = [out],
      visibility = visibility,
      cmd = "cat hack/boilerplate/boilerplate.yaml.txt > \"$@\" && {} >> \"$@\"".format(build_cmd))

  # bazel run <name>.preview to view generated manifest
  native.genrule(
      name = "{}.preview".format(name),
      srcs = [":"+ name],
      outs = ["output"],
      visibility = visibility,
      cmd = "echo 'cat $(location :{})' > \"$@\"".format(name),
      executable = True)
