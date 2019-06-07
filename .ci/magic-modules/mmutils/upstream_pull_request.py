"""Helper class for obtaining information about upstream PR and its downstreams.

  Typical usage example:

  upstream = UpstreamPullRequest(github_token, 100)
  upstream_pr = upstream.fetch()
  downstream_urls = upstream.downstream_urls()
"""

import os
import re
import sys
import itertools
import operator
from github import Github

def find_dependency_urls_in_comment(body):
  """Util to parse downstreams from a given comment body"""
  return re.findall(
    r'^depends: (https://github.com/[^\s]*)', body, re.MULTILINE)

def parse_github_url(self, gh_url):
  """Util to parse Github repo/PR id from a Github PR URL."""
  matches = re.match(r'https://github.com/([\w-]+/[\w-]+)/pull/(\d+)', url)
  return matches.groups() if matches else None

class UpstreamPullRequest(object):
  REPO = 'GoogleCloudPlatform/magic-modules'

  def __init__(self, gh, pr_num):
    """Initializes an instance.

    Args:
      github_token (String): github.Github client object
      pr_num: PR number for the upstream pull request
    """
    self.client = gh
    self.pr_num = pr_num

  def fetch(self):
    """Fetch pull request from Github

    Returns:
      github.PullRequest.PullRequest
    """
    return self.client.get_repo(self.REPO)\
                      .get_pull(self.pr_num)

  def find_unmerged_downstreams(self):
    """Returns list of urls for unmerged, open downstream"""
    unmerged_dependencies = []
    for r, pulls in self.parsed_downstream_urls():
      repo = self.client.get_repo(r)
      for _repo, pr_num in pulls:
        pr = repo.get_pull(pr_num)
        if not pr.is_merged() and not pr.state == "closed":
          unmerged_dependencies.append(pr.html_url)

    return unmerged_dependencies

  def parsed_downstream_urls(self):
    """Get parsed URLs for downstream PRs grouped by repo.

    Usage:
      parsed = UpstreamPullRequest(pr_num).parsed_downstream_urls
      for repo, repo_pulls in parsed:
        for _repo, pr in repo_pulls:
          print "Downstream is https://github.com/%s/pull/%d" % (repo, pr)

    Returns:
      Iterator over $repo and sub-iterators of ($repo, $pr_num) parsed tuples
    """
    parsed = [parse_github_url(u) for u in self.downstream_urls(pr_num)]
    return itertools.groupby(parsed, key=operator.itemgetter(0))

  def downstream_urls(self):
    """Get list of URLs for downstream PRs."""
    urls = []
    for comment in self.fetch().get_issue_comments():
      urls = urls + find_dependency_urls_in_comment(comment.body)
    return urls


