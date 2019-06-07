#!/usr/bin/env python

import os
import sys
from github import Github
from lib.upstream_pull_request import UpstreamPullRequest

if __name__ == '__main__':
  assert len(sys.argv) == 2, "expected id filename as argument"
  with open(sys.argv[1]) as f:
    pr_num = int(f.read())

  upstream = UpstreamPullRequest(Github(os.environ.get('GH_TOKEN')), pr_num)
  unmerged = upstream.get_unmerged_downstreams(pr_num):
  if unmerged:
    raise ValueError("some PRs are unmerged", unmerged)
