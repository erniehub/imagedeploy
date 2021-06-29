ARG HELM_VERSION
ARG KUBERNETES_VERSION
ARG ALPINE_VERSION

FROM "registry.gitlab.com/gitlab-org/cluster-integration/helm-install-image/releases/${HELM_VERSION}-kube-${KUBERNETES_VERSION}-alpine-${ALPINE_VERSION}"

# Install Dependencies
RUN apk add --no-cache \
  bash \
  curl \
  gzip \
  jq \
  openssl \
  ruby \
  ruby-json \
  tar

COPY src/ build/
COPY assets/ assets/

RUN ln -s /build/bin/* /usr/local/bin/
