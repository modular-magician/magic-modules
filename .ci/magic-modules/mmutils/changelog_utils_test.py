#!/usr/bin/env python

from changelog_utils import *
import unittest

class ChangelogUtilTests(unittest.TestCase):
  def test_get_release_note(self):
    upstream_body = """
      ```releasenote
      This is a release note
      ```
    """
    test_cases = {
      ("releasenote text not found", ""),
      ("""
        Empty release note:
        ```releasenote
        ```
        """, ""),
      ("""
        Random code block
        ```
        This is not a release note
        ```
        """, ""),
      ("""
        Empty release note with non-empty code block:
        ```releasenote
        ```

        ```
        This is not a release note
        ```
        """, ""),
      ("""
        Empty code block with non-empty release note:

        ```invalid
        ```

        ```releasenote
        This is a release note
        ```
        """, "This is a release note"),
      ("""```releasenote
        This is a release note
        ```
        """, "This is a release note"),
    }
    for k, v in test_cases:
      self.assertEqual(get_release_note(k), v)

  def test_set_release_note(self):
    downstream_body = """
      All of the blocks below should be replaced

      ```releasenote
      This should be replaced
      ```

      More text

      ```releasenote
      ```

      ```test
      ```
      """
    release_note = "The release note was replaced"

    replaced = set_release_note(release_note, downstream_body)
    self.assertIn("```releasenote\nThe release note was replaced\n```\n", replaced)
    self.assertTrue(len(re.findall("```releasenote", replaced)) == 1)

    self.assertNotIn("This should be replaced", replaced)
    self.assertIn("All of the blocks below should be replaced\n", replaced)
    self.assertIn("More text\n", replaced)

  def test_get_changelog_labels(self):
    self.assertFalse(get_changelog_labels([]))
    self.assertFalse(get_changelog_labels(["", ""]))

    test_labels = [
      "test: foo",
      "test: bar",
      # Not valid changelog labels
      "not a changelog label",
      "test: "
    ]
    result = get_changelog_labels(test_labels, prefix="test: ")

    self.assertTrue(len(result) == 2, "expected only 2 labels returned")
    self.assertIn("test: foo", result)
    self.assertIn("test: bar", result)

  def test_replace_changelog_labels(self):
    original = ["test: baz", "not a changelog label"]
    to_add = ["test: foo", "test: bar"]

    result = replace_changelog_labels(original, to_add, "test: ")
    self.assertTrue(len(result) == 3, "expected only 3 labels")
    self.assertIn("test: foo", result)
    self.assertIn("test: bar", result)
    self.assertIn("not a changelog label", result)
    self.assertNotIn("test: baz", result)

if __name__ == '__main__':
  unittest.main()
