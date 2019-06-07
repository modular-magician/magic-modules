#!/usr/bin/env python
from upstream_pull_request import UpstreamPullRequest
from github import Github
import os
import sys
import argparse
import re

CHANGELOG_LABEL_PREFIX = "changelog: "
CHANGELOG_REPOS = [
  "terraform-providers/terraform-provider-google",
  "terraform-providers/terraform-provider-google-beta",
]

def downstream_changelog_info(gh, prnum):
  upstream = UpstreamPullRequest(gh, prnum)
  upstream_pr = upstream.fetch()
  release_note = parse_release_note(upstream_pr.body)
  labels = [l for l in upstream_pr.get_labels()
              if l.startswith(CHANGELOG_LABEL_PREFIX)]

  if not release_note and not labels:
    print "Upstream PR has no release note or labels, skipping changelog"

  for r, pulls in upstream.parsed_downstream_urls():
    continue if r not in CHANGELOG_REPOS
    repo = gh.get_repo(r)
    for _r, prnum in pulls:
      pr = repo.get_pull(prnum)
      pr.edit(body=overwrite_release_note(pr.body, release_note))
      pr.add_to_labels(labels)

def parse_release_note(pr_body):
  """Util to remove and replace the release note(s) inside a PR body"""
  found = re.search(r'^\n```releasenote\n([.]*)\n```\n)', pr_body)
  if found:
    release_note = re.group(0).strip()
    return release_note
  return ""

def overwrite_release_note(pr_body, release_note):
  """Util to remove and replace the release note(s) inside a PR body"""
  body = re.sub(r'^(`\n``releasenote\n[.]*\n```\n)', '', current_body)
  body += '\n```releasenote\n%s\n```\n'


