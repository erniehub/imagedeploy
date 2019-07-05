#! /bin/sh

export TILLER_NAMESPACE=$KUBE_NAMESPACE

function check_kube_domain() {
  if [[ -z "$KUBE_INGRESS_BASE_DOMAIN" ]]; then
    echo "In order to deploy or use Review Apps,"
    echo "KUBE_INGRESS_BASE_DOMAIN variables must be set"
    echo "From 11.8, you can set KUBE_INGRESS_BASE_DOMAIN in cluster settings"
    echo "or by defining a variable at group or project level."
    echo "You can also manually add it in .gitlab-ci.yml"
    false
  else
    true
  fi
}

function download_chart() {
  if [[ ! -d chart ]]; then
    auto_chart=${AUTO_DEVOPS_CHART:-gitlab/auto-deploy-app}
    auto_chart_name=$(basename $auto_chart)
    auto_chart_name=${auto_chart_name%.tgz}
    auto_chart_name=${auto_chart_name%.tar.gz}
  else
    auto_chart="chart"
    auto_chart_name="chart"
  fi

  helm init --client-only
  helm repo add ${AUTO_DEVOPS_CHART_REPOSITORY_NAME:-gitlab} ${AUTO_DEVOPS_CHART_REPOSITORY:-https://charts.gitlab.io} ${AUTO_DEVOPS_CHART_REPOSITORY_USERNAME:+"--username" "$AUTO_DEVOPS_CHART_REPOSITORY_USERNAME"} ${AUTO_DEVOPS_CHART_REPOSITORY_PASSWORD:+"--password" "$AUTO_DEVOPS_CHART_REPOSITORY_PASSWORD"}
  if [[ ! -d "$auto_chart" ]]; then
    helm fetch ${auto_chart} --untar
  fi
  if [ "$auto_chart_name" != "chart" ]; then
    mv ${auto_chart_name} chart
  fi

  helm dependency update chart/
  helm dependency build chart/
}

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
