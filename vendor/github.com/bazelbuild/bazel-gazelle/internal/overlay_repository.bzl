# Copyright 2018 The Bazel Authors. All rights reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#    http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

def _http_archive_impl(ctx):
  overlay = _resolve_overlay(ctx, ctx.attr.overlay)
  ctx.download_and_extract(
      url = ctx.attr.urls,
      sha256 = ctx.attr.sha256,
      type = ctx.attr.type,
      stripPrefix = ctx.attr.strip_prefix,
  )
  _apply_overlay(ctx, overlay)

http_archive = repository_rule(
    implementation = _http_archive_impl,
    attrs = {
        "urls": attr.string_list(),
        "sha256": attr.string(),
        "strip_prefix": attr.string(),
        "type": attr.string(),
        "overlay": attr.label_keyed_string_dict(allow_files = True),
    },
)
# TODO(jayconrod): add strip_count to remove a number of unnamed
# parent directories.
# TODO(jayconrod): add sha256_contents to check sha256sum on files extracted
# from the archive instead of on the archive itself.

def _git_repository_impl(ctx):
  if not ctx.attr.commit and not ctx.attr.tag:
    fail("either 'commit' or 'tag' must be specified")
  if ctx.attr.commit and ctx.attr.tag:
    fail("'commit' and 'tag' may not both be specified")

  overlay = _resolve_overlay(ctx, ctx.attr.overlay)

  # TODO(jayconrod): sanitize inputs passed to git.
  revision = ctx.attr.commit if ctx.attr.commit else ctx.attr.tag
  _check_execute(ctx, ["git", "clone", "-n", ctx.attr.remote, "."], "failed to clone %s" % ctx.attr.remote)
  _check_execute(ctx, ["git", "checkout", revision], "failed to checkout revision %s in remote %s" % (revision, ctx.attr.remote))
  
  _apply_overlay(ctx, overlay)

git_repository = repository_rule(
    implementation = _git_repository_impl,
    attrs = {
        "commit": attr.string(),
        "remote": attr.string(mandatory = True),
        "tag": attr.string(),
        "overlay": attr.label_keyed_string_dict(allow_files = True),
    },
)

def _resolve_overlay(ctx, overlay):
  """Resolve overlay labels to paths.

  This should be done before downloading the repository, since it may
  trigger restarts.
  """
  return [(ctx.path(src_label), dst_rel) for src_label, dst_rel in overlay.items()]

def _apply_overlay(ctx, overlay):
  """Copies overlay files into the repository.

  This should be done after downloading the repository, since it may replace
  downloaded files.
  """
  # TODO(jayconrod): sanitize destination paths.
  for src_path, dst_rel in overlay:
    _check_execute(ctx, ["cp", src_path, dst_rel], "failed to copy file from %s" % src_path)

def _check_execute(ctx, arguments, message):
  res = ctx.execute(arguments)
  if res.return_code != 0:
    fail(message + "\n" + res.stdout + res.stderr)
