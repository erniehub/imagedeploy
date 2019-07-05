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
