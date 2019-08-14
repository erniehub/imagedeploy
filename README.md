# auto-deploy-image

The [Auto-DevOps](https://docs.gitlab.com/ee/topics/autodevops/) [deploy stage](https://gitlab.com/gitlab-org/gitlab-ce/blob/master/lib/gitlab/ci/templates/Jobs/Deploy.gitlab-ci.yml) image.


## Development

Scripts in this repository follow GitLab's
[shell scripting guide](https://docs.gitlab.com/ee/development/shell_scripting_guide/)
and enforces `shellcheck` and `shfmt`.

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