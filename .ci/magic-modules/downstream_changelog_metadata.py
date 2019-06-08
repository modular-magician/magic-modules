#!/usr/bin/env python
"""
Script to edit downstream PRs with CHANGELOG release note and label metadata.

Usage:
  ./downstream_changelog_info.py path/to/.git/.id
  python /downstream_changelog_info.py

Note that release_note/labels are authoritative - if empty or not set in the MM
upstream PR, release notes will be removed from downstreams and labels
unset.
"""

from mmutils import changelog_utils
from mmutils.upstream_pull_request import UpstreamPullRequest
from github import Github
import os
import argparse

def set_changelog_info(downstream_pull, release_note, changelog_labels):
  """Authoritatively set downstream release note and labels on a downstream PR.

  Args:
    downstream_pull: Downstream PR as github.PullRequest.PullRequest
    release_note: String of release note text to set
    changelog_labels: List of strings changelog labels to set
  """
  print "Setting changelog info for downstream PR %s" % downstream_pull.html_url
  edited_body = changelog_utils.set_release_note(release_note, downstream_pull.body)
  downstream_pull.edit(body=edited_body)

  original_labels = [l.name for l in downstream_pull.get_labels()]
  new_labels = changelog_utils.replace_changelog_labels(
    original_labels, changelog_labels)
  downstream_pull.set_labels(*new_labels)

def downstream_changelog_info(gh, upstream, repos):
  """Edit downstream PRs with CHANGELOG info.

  Args:
    downstream_pull: Downstream PR as github.PullRequest.PullRequest
    release_note: String of release note text to set
    changelog_labels: List of strings changelog labels to set
  """
  # Parse CHANGELOG info from upstream
  upstream_pr = upstream.fetch()
  release_note = changelog_utils.get_release_note(upstream_pr.body)
  original_labels = [l.name for l in upstream_pr.labels]
  changelog_labels = changelog_utils.get_changelog_labels(original_labels)

  print "Applying changelog info to downstreams for upstream PR %d:" % (
    upstream_pr.number)
  print "Release Note: \"%s\"" % release_note
  print "Labels: [%s]" % ", ".join(changelog_labels)

  # Apply to downstreams
  for repo_name, pulls in upstream.parsed_downstream_urls():
    if repo_name not in repos:
      print "[DEBUG] skipping repo %s" % repo_name
      continue

    repo = gh.get_repo(repo_name)
    for _r, pr_num in pulls:
      pr = repo.get_pull(int(pr_num))
      set_changelog_info(pr, release_note, changelog_labels)


if __name__ == '__main__':
  parser = argparse.ArgumentParser()
  parser.add_argument("-fn", "--id_filename", required=True,
    help="File containing number for upstream PR")
  parser.add_argument("-r", "--repos", nargs='+', required=True,
    help="Names of repos containing downstream PRs to copy changelog info to")
  args = parser.parse_args()

  with open(args.id_filename) as f:
    pr_num = int(f.read())

  gh = Github(os.environ.get('GH_TOKEN'))
  upstream = UpstreamPullRequest(gh, pr_num)
  downstream_changelog_info(gh, upstream, args.repos)
