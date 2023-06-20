ARG HELM_INSTALL_IMAGE_VERSION

FROM "registry.gitlab.com/gitlab-org/cluster-integration/helm-install-image:${HELM_INSTALL_IMAGE_VERSION}"

# Magic ARG provided by docker
ARG TARGETARCH

# Install shared dependencies
RUN apk add -u --no-cache \
  bash \
  curl \
  libcurl \
  gzip \
  jq \
  openssl \
  ruby \
  ruby-json \
  tar

# Install libc compatibility pkg using musl
RUN apk add -u --no-cache libc6-compat

COPY src/ build/
COPY assets/ assets/

RUN ln -s /build/bin/* /usr/local/bin/
