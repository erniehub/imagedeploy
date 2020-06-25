# Release

## Semantic Versioning

Auto Deploy Image is versioned according to [Semantic Versioning](https://semver.org/).
Please follow [Specification](https://semver.org/#semantic-versioning-specification-semver)
when you tag a new release.

### Branches

- `master` ... The latest change. Considered as Edge.
- `vX.0.0-pre` ... A pre-release version

## Generating a new auto-deploy image

Please see [the commit rules](development.md#commit-guideline)

### v0.1.0

Starting from GitLab 12.2, the [`Jobs/Deploy.gitlab-ci.yml`](https://gitlab.com/gitlab-org/gitlab-ce/blob/master/lib/gitlab/ci/templates/Jobs/Deploy.gitlab-ci.yml)
template will use the Docker image generated from this project. Changes from previous version of `Jobs/Deploy.gitlab-ci.yml` include:

* Switch from using `sh` to `bash`.
* `install_dependencies` is removed as it is now part of the Docker image.
*  All the other commands should be prepended with `auto-deploy`.
   For example, `check_kube_domain` now becomes `auto-deploy check_kube_domain`.
