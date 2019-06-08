#!/usr/bin/env python
from upstream_pull_request import UpstreamPullRequest
import os
import sys
import argparse
import re

CHANGELOG_LABEL_PREFIX = "changelog: "
RELEASE_NOTE_RE = r'```releasenote[\s]*(?P<release_note>[^`{3}]*)```'

def get_release_note(body):
  """Parse release note block for Terraform changelogs from upstream.

  Finds the first markdown code block with "releasenote" language marker. 
  Example:
    ```releasenote
    This is the release note
    ```

  Args:
    body (string): PR body to pull release note block from

  Returns:
    Release note if found or empty string.
  """
  m = re.search(RELEASE_NOTE_RE, body, re.MULTILINE)
  return m.groupdict("")["release_note"].strip() if m else ""

def set_release_note(release_note, body):
  """Sanitize and add release note block for Terraform changelogs in PR body.

  Example:
    # Set a release note
    > print set_release_note(
        "This is the new release note",
        "``releasenote\nChanges to downstream\n```\n")
    "```releasenote\nThis is the new release note\n```\n"

    # Remove for empty release note
    > print set_release_note("",
        "PR description\n```releasenote\nChanges to downstream\n```\n")
    "PR description\n"

  Args:
    release_note (string): Release note to add
    body (string): PR body to place release note block in

  Returns:
    Text with modified PR body
  """
  edited = re.sub(RELEASE_NOTE_RE, '', body)
  release_note = release_note.strip()
  if release_note:
    edited += "\n```releasenote\n%s\n```\n" % release_note.strip()
  return edited

def get_changelog_labels(labels, prefix=CHANGELOG_LABEL_PREFIX):
  """Find labels on upstream PR that should be transferred to downstream PRs.

  Args:
    prefix: String expected to be prefix of CHANGELOG-relevant labels
    labels: List of string labels from upstream PR

  Return:
    String list of CHANGELOG-relevant label names (with prefix removed)
  """
  changelog_labels = []
  for l in labels:
    if not l.startswith(prefix):
      continue

    label = l[len(prefix):].strip()
    if label:
      changelog_labels.append(prefix + label)

  return changelog_labels

def replace_changelog_labels(original, to_add, prefix=CHANGELOG_LABEL_PREFIX):
  """Replace CHANGELOG-relevant labels in downstream PR labels.

  Args:
    prefix: String expected to be prefix of CHANGELOG-relevant labels
    original: Current labels on downstream PR as list of strings
    to_add: List of CHANGELOG-relevant label names to add

  Return:
    List of new labels to set
  """
  return [l for l in original if not l.startswith(prefix)] + to_add
