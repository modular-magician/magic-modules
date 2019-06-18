#!/usr/bin/env python

import os
import sys
from github import Github
from mmutils.upstream_pull_request import UpstreamPullRequest

if __name__ == '__main__':
  assert len(sys.argv) == 2, "expected a Github PR ID as argument"
  pr_num = int(sys.argv[1])

  upstream = UpstreamPullRequest(Github(os.environ.get('GH_TOKEN')), pr_num)
  for url in upstream.downstream_urls():
    print url
