#! /bin/sh

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
