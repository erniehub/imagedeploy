# API

auto-deploy-image provides the following APIs to orchestrate [GitLab Auto Deploy](https://docs.gitlab.com/ce/topics/autodevops/stages.html#auto-deploy).

## Common arguments for all APIs

| Arguments           | Type                           | Required | Description | Available |
|---------------------|--------------------------------|----------|-------------|-------------|
| `CI_COMMIT_REF_SLUG`                   | string | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `CI_COMMIT_TAG`                        | string | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `CI_ENVIRONMENT_NAME`                  | string | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `CI_ENVIRONMENT_SLUG`                  | string | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `CI_ENVIRONMENT_URL`                   | string | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `CI_PROJECT_ID`                        | integer | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `CI_PROJECT_PATH_SLUG`                 | string | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `CI_PROJECT_VISIBILITY`                | string | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `CI_REGISTRY_IMAGE`                    | string | yes       | See [GitLab CI Predefined Variables](https://docs.gitlab.com/ee/ci/variables/predefined_variables.html). | v0.1.0 ~ |
| `KUBE_CONTEXT`                         | string | no        | Context to use from within `KUBECONFIG` | v2.16.0 ~ |
| `KUBE_INGRESS_BASE_DOMAIN`             | string | yes       | See [GitLab Cluster Integration Deployment Variables](https://docs.gitlab.com/ee/user/project/clusters/). | v0.1.0 ~ |
| `KUBE_NAMESPACE`                       | string | no        | The deployment namespace. If not specified, the context default will be used. If the context has no default, falls back to `default` | v0.1.0 ~ |
| `KUBECONFIG`                           | string | yes       | See [GitLab Cluster Integration Deployment Variables](https://docs.gitlab.com/ee/user/project/clusters/). | v0.1.0 ~ |
| `AUTO_DEVOPS_DEPLOY_DEBUG`             | boolean | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | [v0.16.0](https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/compare/v0.15.0...v0.16.0) ~ |
| `HELM_RELEASE_NAME`                    | string | no        | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |

## Check the base domain for ingress

> **Notes**:
>
> - Introduced in auto-deploy-image v0.1.0.

Example:

```shell
auto-deploy check_kube_domain
```

## Ensure Helm chart existence

> **Notes**:
>
> - Starting from v1.0.0, a chart in an asset directory is used instead of downloading it from charts.gitlab.io.
> - Introduced in auto-deploy-image v0.1.0.

Ensure the existence of a helm chart for [deployment](#deploy).
If `chart` directory doesn't exist in the current location, it places a
auto-deploy-app chart from `assets` directory or `charts.gitlab.io` to the `chart` directory.
Alternatively, you can specifiy environment variables to fetch chart from a
specific chart repository.

| Arguments           | Type                           | Required | Description | Available |
|---------------------|--------------------------------|----------|-------------|-------------|
| `AUTO_DEVOPS_CHART_REPOSITORY_NAME`       | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `AUTO_DEVOPS_CHART_REPOSITORY_PASSWORD`   | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `AUTO_DEVOPS_CHART_REPOSITORY_USERNAME`   | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `AUTO_DEVOPS_CHART_REPOSITORY`            | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `AUTO_DEVOPS_CHART`                       | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |

Example:

```shell
auto-deploy download_chart
```

## Ensuring kubernetes namespace existence

> **Notes**:
>
> - Introduced in auto-deploy-image v0.1.0.

Ensure the existence of namespace for [deployment](#deploy).
It creates a new namespace if it doesn't exist yet.

Example:

```shell
auto-deploy ensure_namespace
```

## Initialize Tiller

> **Notes**:
>
> - Introduced in auto-deploy-image v0.1.0.
> - Removed in auto-deploy-image v2.0.0, as part of upgrading to Helm 3

Example:

```shell
auto-deploy initialize_tiller
```

## Create a secret

> **Notes**:
>
> - Introduced in auto-deploy-image v0.1.0.

Create a secret for allowing the cluster to pull an application image from a private project.

Example:

```shell
auto-deploy create_secret
```

## Deploy

> **Notes**:
>
> - Introduced in auto-deploy-image v0.1.0.

Deploy an application

| Arguments           | Type                           | Required | Description | Available |
|---------------------|--------------------------------|----------|-------------|-------------|
| 1st argument                                  | string | no        | The release track. One of `stable`, `canary` or `rollout`. Default is `stable`. | v0.1.0 ~ |
| 2nd argument                                  | integer | no       | The percentage of rollout. Default is `100`. | v0.1.0 ~ |
| `<ENVIRONMENT>_ADDITIONAL_HOSTS`              | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `AUTO_DEVOPS_ALLOW_TO_FORCE_DEPLOY_V<N>`      | boolean | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v1.0.0 ~ |
| `AUTO_DEVOPS_ATOMIC_RELEASE`                  | integer | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | [v0.13.1](https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/compare/v0.13.0...v0.13.1) ~ |
| `AUTO_DEVOPS_MODSECURITY_SEC_RULE_ENGINE`     | integer | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | [v0.3.0](https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/compare/v0.2.2...v0.3.0) ~ |
| `AUTO_DEVOPS_POSTGRES_CHANNEL`                | integer | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | [v0.12.0](https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/compare/v0.11.0...v0.12.0) ~ v2.0.0 |
| `AUTO_DEVOPS_POSTGRES_DELETE_V1`              | integer | no       | See [Upgrading PostgreSQL](https://docs.gitlab.com/ee/topics/autodevops/upgrading_postgresql.html). | [v0.13.3](https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/compare/v0.13.2...v0.13.3) ~ v2.0.0 |
| `AUTO_DEVOPS_POSTGRES_MANAGED_CLASS_SELECTOR` | integer | no       | See [Crossplane configuration](https://docs.gitlab.com/ee/user/clusters/crossplane.html). | [v0.7.0](https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/compare/v0.6.0...v0.7.0) ~ |
| `AUTO_DEVOPS_POSTGRES_MANAGED`                | string | no       | See [Crossplane configuration](https://docs.gitlab.com/ee/user/clusters/crossplane.html). | [v0.7.0](https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/compare/v0.6.0...v0.7.0) ~ |
| `CI_APPLICATION_REPOSITORY`                   | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `CI_APPLICATION_TAG`                          | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `DB_INITIALIZE`                               | boolean | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `HELM_UPGRADE_EXTRA_ARGS`                     | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `HELM_UPGRADE_VALUES_FILE`                    | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | [v0.8.0](https://gitlab.com/gitlab-org/cluster-integration/auto-deploy-image/compare/v0.7.0...v0.8.0) ~ |
| `POSTGRES_DB`                                 | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `POSTGRES_ENABLED`                            | boolean | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `POSTGRES_PASSWORD`                           | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `POSTGRES_USER`                               | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `POSTGRES_VERSION`                            | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `POSTGRES_HELM_UPGRADE_EXTRA_ARGS`            | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v2.0.2 ~ |
| `POSTGRES_HELM_UPGRADE_VALUES_FILE`           | string | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v2.0.2 ~ |
| `ROLLOUT_RESOURCE_TYPE`                       | integer | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |
| `ROLLOUT_STATUS_DISABLED`                     | boolean | no       | See [Customizing Auto DevOps](https://docs.gitlab.com/ee/topics/autodevops/customize.html). | v0.1.0 ~ |

Example:

```shell
auto-deploy deploy canary
```

## Scale up or down pods

> **Notes**:
>
> - Introduced in auto-deploy-image v0.1.0.

| Arguments           | Type                           | Required | Description | Available |
|---------------------|--------------------------------|----------|-------------|-------------|
| 1st argument        | string | no        | The release track. One of `stable`, `canary` or `rollout`. Default is `stable`. | v0.1.0 ~ |
| 2nd argument        | integer | no       | The percentage of rollout. Default is `100`. | v0.1.0 ~ |

Example:

```shell
auto-deploy scale
```

## Delete an environment

> **Notes**:
>
> - Introduced in auto-deploy-image v0.1.0.

| Arguments           | Type                           | Required | Description | Available |
|---------------------|--------------------------------|----------|-------------|-------------|
| 1st argument        | string | no        | The release track. One of `stable`, `canary` or `rollout`. Default is `stable`. | v0.1.0 ~ |

Example:

```shell
auto-deploy delete canary
```

### Persist the URL of a created environment

> **Notes**:
>
> - Introduced in auto-deploy-image v0.1.0.

Example:

```shell
auto-deploy persist_environment_url
```
