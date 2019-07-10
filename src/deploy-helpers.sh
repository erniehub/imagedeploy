#! /bin/sh
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
