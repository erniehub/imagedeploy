# auto-deploy-image

The [Auto-DevOps](https://docs.gitlab.com/ee/topics/autodevops/) [deploy stage](https://gitlab.com/gitlab-org/gitlab-ce/blob/master/lib/gitlab/ci/templates/Jobs/Deploy.gitlab-ci.yml) image.

## Notice

We're moving [auto-deploy-app chart](https://gitlab.com/gitlab-org/charts/auto-deploy-app) into this project
in order for [Versioning charts in order to ship breaking change safely](https://gitlab.com/gitlab-org/charts/auto-deploy-app/-/issues/70),
you see bundled charts in `vendor/auto-deploy-app-chart`.

## Development

Read about the [Development guide](doc/development.md)

## Contributing and Code of Conduct

Please see [CONTRIBUTING.md](CONTRIBUTING.md)

## Release

Read about the [Package release process](doc/release.md)
