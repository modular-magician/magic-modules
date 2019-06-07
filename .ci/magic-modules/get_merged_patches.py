#!/usr/bin/env python
import os
import urllib
from github import Github
from lib.upstream_pull_request import UpstreamPullRequest

def get_merged_patches(gh):
  """Download all merged patches for open upstream PRs.

  Args:
    gh: Github handle to make calls to Github with.
  """
  open_pulls = gh.get_repo('GoogleCloudPlatform/magic-modules')\
                 .get_pulls(state='open')
  for open_pr in open_pulls:
    upstream = UpstreamPullRequest(gh, open_pr.number)

    print 'Downloading patches for upstream PR %d\n...' % open_pr.number
    for repo_name, pulls in upstream.parsed_downstream_urls():
      repo = gh.get_repo(repo_name)
      for r, pr_id in pulls:
          print 'Check to see if %s/%s is merged and should be downloaded\n' % (
            r, pr_id)
          downstream_pr = repo.get_pull(int(pr_id))
          if downstream_pr.is_merged():
            download_patch(r, downstream_pr)

def download_patch(repo, pr):
  """Download merged downstream PR patch.

  Args:
    pr: Github Pull request to download patch for
  """
  download_location = os.path.join('./patches', repo_name, '%d.patch' % pr.id)

  # Skip already downloaded patches
  if os.path.exists(download_location):
    return

  if not os.path.exists(os.path.dirname(download_location)):
      os.makedirs(os.path.dirname(download_location))
  urllib.urlretrieve(pr.patch_url, download_location)

if __name__ == '__main__':
  gh = Github(os.environ.get('GH_TOKEN'))
  get_merged_patches(gh)
