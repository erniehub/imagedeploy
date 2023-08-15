# auto-deploy-image

The [Auto-DevOps](https://docs.gitlab.com/ee/topics/autodevops/) [deploy stage](https://gitlab.com/gitlab-org/gitlab/-/blob/master/lib/gitlab/ci/templates/Jobs/Deploy.gitlab-ci.yml) image.

## API

Please see [API documentation](doc/api.md)

## Development

Scripts in this repository follow GitLab's
[shell scripting guide](https://docs.gitlab.com/ee/development/shell_scripting_guide/)
and enforces `shellcheck` and `shfmt`.

## Testing

When choosing what to test, for example in`.gitlab/ci/test.gitlab-ci.yml`, test GitLab's [currently supported Kubernetes versions](https://docs.gitlab.com/ee/user/clusters/agent/#supported-cluster-versions).

## Contributing and Code of Conduct

Please see [CONTRIBUTING.md](CONTRIBUTING.md)

## Upgrading

### v0.1.0

Starting from GitLab 12.2, the [`Jobs/Deploy.gitlab-ci.yml`](https://gitlab.com/gitlab-org/gitlab-ce/blob/master/lib/gitlab/ci/templates/Jobs/Deploy.gitlab-ci.yml)
template will use the Docker image generated from this project. Changes from previous version of `Jobs/Deploy.gitlab-ci.yml` include:

* Switch from using `sh` to `bash`.
* `install_dependencies` is removed as it is now part of the Docker image.
*  All the other commands should be prepended with `auto-deploy`.
   For example, `check_kube_domain` now becomes `auto-deploy check_kube_domain`.

### v1.0.0

Before v1.0.0, auto-deploy-image was downloading a chart from [the chart repository](https://charts.gitlab.io/),
which was then uploaded by the [auto-deploy-app](https://gitlab.com/gitlab-org/charts/auto-deploy-app) project.

Since auto-deploy-image v1.0.0, the auto-deploy-app chart is bundled into the auto-deploy-image docker image as a local asset,
and it no longer downloads the chart from the repository.

# Generating a new auto-deploy image

To generate a new image you must follow the git commit guidelines below, this
will trigger a semantic version bump which will then cause a new pipeline
that will build and tag the new image

## Git Commit Guidelines

This project uses [Semantic Versioning](https://semver.org). We use commit
messages to automatically determine the version bumps, so they should adhere to
the conventions of [Conventional Commits (v1.0.0)](https://www.conventionalcommits.org/en/v1.0.0).

### TL;DR

- Commit title starting with `fix: ` trigger a patch version bump
- Commit title starting with `feat: ` trigger a minor version bump
- Commit body contains `BREAKING CHANGE: ` trigger a major version bump. This can be part of commits of any _type_.

### Tip: Test commit messages locally

Testing the commit message locally can speed up the iteration cycle. It can be configured as follows:

``` sh
# install dev dependencies, if necessary
npm install

# usage
npx commitlint --from=master # if targeting latest
npx commitlint --from=1.x # if targeting 1.x stable
```

### Tip: Use a git hook with commitlint

To save yourself the manual step of testing the commit message, you can use a commit hook.

At the root of this project, add `.git/hooks/commit-msg` with the following contents:

``` sh
#!/bin/sh
npx commitlint --edit
```

Then, run `chmod +x .git/hooks/commit-msg` to make it executable.

## Automatic versioning

Each push to `master` triggers a [`semantic-release`](https://semantic-release.gitbook.io/semantic-release/)
CI job that determines and pushes a new version tag (if any) based on the
last version tagged and the new commits pushed. Notice that this means that if a
Merge Request contains, for example, several `feat: ` commits, only one minor
version bump will occur on merge. If your Merge Request includes several commits
you may prefer to ignore the prefix on each individual commit and instead add
an empty commit summarizing your changes like so:

```
git commit --allow-empty -m '[BREAKING CHANGE|feat|fix]: <changelog summary message>'
```

## Backport

For backporting a change to a previous release, please follow [this recipe](https://github.com/semantic-release/semantic-release/blob/master/docs/recipes/maintenance-releases.md#publishing-maintenance-releases).

Here is an example:

Given you have

- `v2.0.0` tag, which is the latest tag of `v2`.
- `v1.1.1` tag, which is the latest tag of `v1`.
- `master` branch, which points to `v2.0.0`.

and you want to release a bug fix to both `v1` and `v2` releases.

You process the following actions:

- Create `1.x` branch from `v1.1.1`.
- Merge a fix to `1.x` branch. `semantic-release` creates a release with  `v1.1.2` tag.
- Merge a fix to `master` branch. `semantic-release` creates a release with `v2.0.1` tag.

NOTE: **NOTE**
Ensure that the maintenance release branch (e.g. `1.x`) is [protected](https://docs.gitlab.com/ee/user/project/protected_branches.html).

## Pre-release

For publishing a pre-release, please follow [this recipe](https://github.com/semantic-release/semantic-release/blob/master/docs/recipes/pre-releases.md#publishing-pre-releases).

Here is an example:

Given you have

- `v1.1.1` tag, which is the latest tag of `v1`.
- `master` branch, which points to `v1.1.1`.

and you want to publish a pre-release.

You process the following actions:

- Create `beta` branch from the latest `master`.
- Merge a fix to `beta` branch. `semantic-release` creates a release with `2.0.0-beta.1` tag.
- When you make an official release, merge `beta` to `master`. `semantic-release` creates a release with `2.0.0` tag.
- Delete the `beta` branch on https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/-/branches

NOTE: **NOTE**
Ensure that the pre-release release branch (e.g. `beta`) is [protected](https://docs.gitlab.com/ee/user/project/protected_branches.html).
