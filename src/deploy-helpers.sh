#! /bin/sh

[[ "$TRACE" ]] && set -x

auto_database_url=postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${CI_ENVIRONMENT_SLUG}-postgres:5432/${POSTGRES_DB}
export DATABASE_URL=${DATABASE_URL-$auto_database_url}
export TILLER_NAMESPACE=$KUBE_NAMESPACE

function ensure_namespace() {
  kubectl get namespace "$KUBE_NAMESPACE" || kubectl create namespace "$KUBE_NAMESPACE"
}

function initialize_tiller() {
  echo "Checking Tiller..."

  export HELM_HOST="localhost:44134"
  tiller -listen ${HELM_HOST} -alsologtostderr > /dev/null 2>&1 &
  echo "Tiller is listening on ${HELM_HOST}"

  if ! helm version --debug; then
    echo "Failed to init Tiller."
    return 1
  fi
  echo ""
}

function create_secret() {
  echo "Create secret..."
  if [[ "$CI_PROJECT_VISIBILITY" == "public" ]]; then
    return
  fi

  kubectl create secret -n "$KUBE_NAMESPACE" \
    docker-registry gitlab-registry \
    --docker-server="$CI_REGISTRY" \
    --docker-username="${CI_DEPLOY_USER:-$CI_REGISTRY_USER}" \
    --docker-password="${CI_DEPLOY_PASSWORD:-$CI_REGISTRY_PASSWORD}" \
    --docker-email="$GITLAB_USER_EMAIL" \
    -o yaml --dry-run | kubectl replace -n "$KUBE_NAMESPACE" --force -f -
}

function persist_environment_url() {
  echo $CI_ENVIRONMENT_URL > environment_url.txt
}

function deploy() {
  track="${1-stable}"
  percentage="${2:-100}"
  name=$(deploy_name "$track")

  if [[ -z "$CI_COMMIT_TAG" ]]; then
    image_repository=${CI_APPLICATION_REPOSITORY:-$CI_REGISTRY_IMAGE/$CI_COMMIT_REF_SLUG}
    image_tag=${CI_APPLICATION_TAG:-$CI_COMMIT_SHA}
  else
    image_repository=${CI_APPLICATION_REPOSITORY:-$CI_REGISTRY_IMAGE}
    image_tag=${CI_APPLICATION_TAG:-$CI_COMMIT_TAG}
  fi

  service_enabled="true"
  postgres_enabled="$POSTGRES_ENABLED"

  # if track is different than stable,
  # re-use all attached resources
  if [[ "$track" != "stable" ]]; then
    service_enabled="false"
    postgres_enabled="false"
  fi

  replicas=$(get_replicas "$track" "$percentage")

  if [[ "$CI_PROJECT_VISIBILITY" != "public" ]]; then
    secret_name='gitlab-registry'
  else
    secret_name=''
  fi

  create_application_secret "$track"

  env_slug=$(echo ${CI_ENVIRONMENT_SLUG//-/_} | tr -s '[:lower:]' '[:upper:]')
  eval env_ADDITIONAL_HOSTS=\$${env_slug}_ADDITIONAL_HOSTS
  if [ -n "$env_ADDITIONAL_HOSTS" ]; then
    additional_hosts="{$env_ADDITIONAL_HOSTS}"
  elif [ -n "$ADDITIONAL_HOSTS" ]; then
    additional_hosts="{$ADDITIONAL_HOSTS}"
  fi

  if [[ -n "$DB_INITIALIZE" && -z "$(helm ls -q "^$name$")" ]]; then
    echo "Deploying first release with database initialization..."
    helm upgrade --install \
      --wait \
      --set service.enabled="$service_enabled" \
      --set gitlab.app="$CI_PROJECT_PATH_SLUG" \
      --set gitlab.env="$CI_ENVIRONMENT_SLUG" \
      --set releaseOverride="$CI_ENVIRONMENT_SLUG" \
      --set image.repository="$image_repository" \
      --set image.tag="$image_tag" \
      --set image.pullPolicy=IfNotPresent \
      --set image.secrets[0].name="$secret_name" \
      --set application.track="$track" \
      --set application.database_url="$DATABASE_URL" \
      --set application.secretName="$APPLICATION_SECRET_NAME" \
      --set application.secretChecksum="$APPLICATION_SECRET_CHECKSUM" \
      --set service.commonName="le-$CI_PROJECT_ID.$KUBE_INGRESS_BASE_DOMAIN" \
      --set service.url="$CI_ENVIRONMENT_URL" \
      --set service.additionalHosts="$additional_hosts" \
      --set replicaCount="$replicas" \
      --set postgresql.enabled="$postgres_enabled" \
      --set postgresql.nameOverride="postgres" \
      --set postgresql.postgresUser="$POSTGRES_USER" \
      --set postgresql.postgresPassword="$POSTGRES_PASSWORD" \
      --set postgresql.postgresDatabase="$POSTGRES_DB" \
      --set postgresql.imageTag="$POSTGRES_VERSION" \
      --set application.initializeCommand="$DB_INITIALIZE" \
      $HELM_UPGRADE_EXTRA_ARGS \
      --namespace="$KUBE_NAMESPACE" \
      "$name" \
      chart/

    echo "Deploying second release..."
    helm upgrade --reuse-values \
      --wait \
      --set application.initializeCommand="" \
      --set application.migrateCommand="$DB_MIGRATE" \
      $HELM_UPGRADE_EXTRA_ARGS \
      --namespace="$KUBE_NAMESPACE" \
      "$name" \
      chart/
  else
    echo "Deploying new release..."
    helm upgrade --install \
      --wait \
      --set service.enabled="$service_enabled" \
      --set gitlab.app="$CI_PROJECT_PATH_SLUG" \
      --set gitlab.env="$CI_ENVIRONMENT_SLUG" \
      --set releaseOverride="$CI_ENVIRONMENT_SLUG" \
      --set image.repository="$image_repository" \
      --set image.tag="$image_tag" \
      --set image.pullPolicy=IfNotPresent \
      --set image.secrets[0].name="$secret_name" \
      --set application.track="$track" \
      --set application.database_url="$DATABASE_URL" \
      --set application.secretName="$APPLICATION_SECRET_NAME" \
      --set application.secretChecksum="$APPLICATION_SECRET_CHECKSUM" \
      --set service.commonName="le-$CI_PROJECT_ID.$KUBE_INGRESS_BASE_DOMAIN" \
      --set service.url="$CI_ENVIRONMENT_URL" \
      --set service.additionalHosts="$additional_hosts" \
      --set replicaCount="$replicas" \
      --set postgresql.enabled="$postgres_enabled" \
      --set postgresql.nameOverride="postgres" \
      --set postgresql.postgresUser="$POSTGRES_USER" \
      --set postgresql.postgresPassword="$POSTGRES_PASSWORD" \
      --set postgresql.postgresDatabase="$POSTGRES_DB" \
      --set postgresql.imageTag="$POSTGRES_VERSION" \
      --set application.migrateCommand="$DB_MIGRATE" \
      $HELM_UPGRADE_EXTRA_ARGS \
      --namespace="$KUBE_NAMESPACE" \
      "$name" \
      chart/
  fi

  if [[ -z "$ROLLOUT_STATUS_DISABLED" ]]; then
    kubectl rollout status -n "$KUBE_NAMESPACE" -w "$ROLLOUT_RESOURCE_TYPE/$name"
  fi
}

function scale() {
  track="${1-stable}"
  percentage="${2-100}"
  name=$(deploy_name "$track")

  replicas=$(get_replicas "$track" "$percentage")

  if [[ -n "$(helm ls -q "^$name$")" ]]; then
    helm upgrade --reuse-values \
      --wait \
      --set replicaCount="$replicas" \
      --namespace="$KUBE_NAMESPACE" \
      "$name" \
      chart/
  fi
}

function delete() {
  track="${1-stable}"
  name=$(deploy_name "$track")

  if [[ -n "$(helm ls -q "^$name$")" ]]; then
    helm delete --purge "$name"
  fi

  secret_name=$(application_secret_name "$track")
  kubectl delete secret --ignore-not-found -n "$KUBE_NAMESPACE" "$secret_name"
}

## Helper functions

# Extracts variables prefixed with K8S_SECRET_
# and creates a Kubernetes secret.
#
# e.g. If we have the following environment variables:
#   K8S_SECRET_A=value1
#   K8S_SECRET_B=multi\ word\ value
#
# Then we will create a secret with the following key-value pairs:
#   data:
#     A: dmFsdWUxCg==
#     B: bXVsdGkgd29yZCB2YWx1ZQo=
function create_application_secret() {
  track="${1-stable}"
  export APPLICATION_SECRET_NAME=$(application_secret_name "$track")

  env | sed -n "s/^K8S_SECRET_\(.*\)$/\1/p" > k8s_prefixed_variables

  kubectl create secret \
    -n "$KUBE_NAMESPACE" generic "$APPLICATION_SECRET_NAME" \
    --from-env-file k8s_prefixed_variables -o yaml --dry-run |
    kubectl replace -n "$KUBE_NAMESPACE" --force -f -

  export APPLICATION_SECRET_CHECKSUM=$(cat k8s_prefixed_variables | sha256sum | cut -d ' ' -f 1)

  rm k8s_prefixed_variables
}

function application_secret_name() {
  track="${1-stable}"
  name=$(deploy_name "$track")

  echo "${name}-secret"
}

function deploy_name() {
  name="$CI_ENVIRONMENT_SLUG"
  track="${1-stable}"

  if [[ "$track" != "stable" ]]; then
    name="$name-$track"
  fi

  echo $name
}

function get_replicas() {
  track="${1:-stable}"
  percentage="${2:-100}"

  env_track=$( echo $track | tr -s  '[:lower:]'  '[:upper:]' )
  env_slug=$( echo ${CI_ENVIRONMENT_SLUG//-/_} | tr -s  '[:lower:]'  '[:upper:]' )

  if [[ "$track" == "stable" ]] || [[ "$track" == "rollout" ]]; then
    # for stable track get number of replicas from `PRODUCTION_REPLICAS`
    eval new_replicas=\$${env_slug}_REPLICAS
    if [[ -z "$new_replicas" ]]; then
      new_replicas=$REPLICAS
    fi
  else
    # for all tracks get number of replicas from `CANARY_PRODUCTION_REPLICAS`
    eval new_replicas=\$${env_track}_${env_slug}_REPLICAS
    if [[ -z "$new_replicas" ]]; then
      eval new_replicas=\${env_track}_REPLICAS
    fi
  fi

  replicas="${new_replicas:-1}"
  replicas="$(($replicas * $percentage / 100))"

  # always return at least one replicas
  if [[ $replicas -gt 0 ]]; then
    echo "$replicas"
  else
    echo 1
  fi
}
