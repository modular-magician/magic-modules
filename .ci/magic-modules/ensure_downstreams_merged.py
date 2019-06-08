#!/usr/bin/env python

import os
import sys
from github import Github
from mmutils.upstream_pull_request import UpstreamPullRequest

if __name__ == '__main__':
  assert len(sys.argv) == 2, "expected id filename as argument"
  with open(sys.argv[1]) as f:
    pr_num = int(f.read())

  upstream = UpstreamPullRequest(Github(os.environ.get('GH_TOKEN')), pr_num)
  unmerged = upstream.find_unmerged_downstreams()
  if unmerged:
    raise ValueError("some PRs are unmerged", unmerged)
