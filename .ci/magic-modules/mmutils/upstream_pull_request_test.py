from upstream_pull_request import *
import unittest
import os
from github import Github

class TestUpstreamPullRequests(unittest.TestCase):
  """
    Terrible test data from scraping
    https://github.com/GoogleCloudPlatform/magic-modules/pull/1000
    TODO: If this test becomes load-bearing, mock out the Github client instead
    of using this.
  """
  TEST_PR_NUM = 1000

  EXPECTED_DOWNSTREAM_URLS = [
    "https://github.com/terraform-providers/terraform-provider-google-beta/pull/186",
    "https://github.com/terraform-providers/terraform-provider-google/pull/2591",
    "https://github.com/modular-magician/ansible/pull/142",
  ]

  EXPECTED_PARSED_DOWNSTREAMS = {
    "terraform-providers/terraform-provider-google-beta": [186],
    "terraform-providers/terraform-provider-google": [2591],
    "modular-magician/ansible": [142],
  }

  def test_find_unmerged_downstreams(self):
    upstream = UpstreamPullRequest(self.github_test_client(), self.TEST_PR_NUM)
    self.assertFalse(upstream.find_unmerged_downstreams())

  def test_parsed_downstream_urls(self):
    upstream = UpstreamPullRequest(self.github_test_client(), self.TEST_PR_NUM)
    result = upstream.parsed_downstream_urls()

    repo_cnt = 0
    for repo, pulls in result:
      # Verify each repo in result.
      self.assertIn(repo, self.EXPECTED_PARSED_DOWNSTREAMS,
        "unexpected repo %s in result" % repo)
      repo_cnt += 1

      # Verify each pull request in result.
      expected_pulls = self.EXPECTED_PARSED_DOWNSTREAMS[repo]
      pull_cnt = 0
      for repo, prid in pulls:
        self.assertIn(int(prid), expected_pulls)
        pull_cnt += 1
      # Verify exact count of pulls (here because iterator).
      self.assertEquals(pull_cnt, len(expected_pulls),
        "expected %d pull requests in result[%s]" % (len(expected_pulls), repo))

    # Verify exact count of repos (here because iterator).
    self.assertEquals(repo_cnt, len(self.EXPECTED_PARSED_DOWNSTREAMS),
        "expected %d pull requests in result[%s]" % (
          len(self.EXPECTED_PARSED_DOWNSTREAMS), repo))

  def test_downstream_urls(self):
    upstream = UpstreamPullRequest(self.github_test_client(), self.TEST_PR_NUM)
    result = upstream.downstream_urls()

    expected_cnt = len(self.EXPECTED_DOWNSTREAM_URLS)
    self.assertEquals(len(result), expected_cnt,
      "expected %d downstream urls, got %d" % (expected_cnt, len(result)))
    for url in result:
      self.assertIn(str(url), self.EXPECTED_DOWNSTREAM_URLS)

  def test_find_dependency_urls(self):
    test_urls = [
      "https://github.com/repo-owner/repo-A/pull/1",
      "https://github.com/repo-owner/repo-A/pull/2",
      "https://github.com/repo-owner/repo-B/pull/3",
    ]
    test_body = "".join(["\ndepends: %s\n" % u for u in test_urls])
    result = find_dependency_urls_in_comment(test_body)
    self.assertEquals(len(result), len(test_urls),
      "expected %d urls to be parsed from comment" % len(test_urls))
    for test_url in test_urls:
      self.assertIn(test_url, result)

  def test_parse_github_url(self):
    test_cases = {
      "https://github.com/repoowner/reponame/pull/1234": ("repoowner/reponame", 1234),
      "not a real url": None,
    }
    for k in test_cases:
      result = parse_github_url(k)
      expected = test_cases[k]
      if not expected:
        self.assertIsNone(result, "expected None, got %s" % result)
      else:
        self.assertEquals(result[0], expected[0])
        self.assertEquals(int(result[1]), expected[1])

  def github_test_client(self):
    ghtoken = os.environ.get('TEST_GITHUB_TOKEN')
    if ghtoken == "":
      self.skip_test("skipping test, env var GITHUB_TOKEN not set")
    return Github(ghtoken)

if __name__ == '__main__':
    unittest.main()


