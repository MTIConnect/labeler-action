# Pull Request Labeler

A [GitHub Action](https://help.github.com/en/categories/automating-your-workflow-with-github-actions)
that labels Pull Requests based on configurable conditions.

Inspired by [srvaroa's Condition based Pull Request Labeler](https://github.com/srvaroa/labeler),
but focused for our workflow at MTI.

## Configuration

Create a `.github/workflows/main.yml` file in the target repository if one doens't already exist:
```yaml
name: CI

on:
  pull_request:
      types: [opened, synchronize, reopened, ready_for_review]
  pull_request_review: {}

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - name: "Labeler"
      uses: MTIConnect/labeler-action@master
      with:
        GITHUB_TOKEN: "${{ secrets.GITHUB_TOKEN }}"
```

Then create config file `.github/pr-labeler.yml` on the default branch, for example:
```yaml
WIP:
  draft: true

Bug:
  branch_name: "^(bug|issue)/"

Feature:
  branch_name: "^(feature|enhancement)/"

Awaiting Code Review:
  draft: false
  changes_requested: false
  approved: false

Changes Requested:
  changes_requested: true

Code Review Approved:
  changes_requested: false
  approved: true
```


## Configuration

By default the configuration for the action is `./github/pr-labeler.yml`.
Where each root key is the label you want to manage. Under each is a map of
conditions that are evaluated. If all conditions of the label evaluate to true,
then the label is added. Otherwise, the label is removed. For example:

```
# The label to add/remove on the following conditions.
Awaiting Code Review:

  # The pull request is not a draft.
  draft: false

  # No reviews that requested changes.
  changes_requested: false

  # No reviews approved the PR.
  approved: false
```

When a non-draft is now opened on the repo, it would apply the `Awaiting Code Review`
label, because there have been no reviews and it is not a draft. However, once the PR
was reviewed with either `Changes Requested` or `Approved` it would be removed.

## Implemented Conditions

### Title

Accepts regex to test against the title of the PR.

```yaml
Refactor:
  title: "^Refactor"
```

### Branch Name

Accepts regex to test against the name of the branch.

```yaml
Bug:
  branch_name: "^(bug|issue)/"
```

### Draft

Accepts a boolean value that tests against if the branch is a [draft](https://github.blog/2019-02-14-introducing-draft-pull-requests/) or not.

```yaml
WIP:
  draft: true
```

### Approved

Accepts a boolean value that tests if there are *any* reviews that approve the changes.

```yaml
Approved:
  approved: true
```

### Changes Requested

Accepts a boolean value that tests if there are *any* reviews requested changes.

```yaml
Changes Requested:
  changes_requested: true
```
