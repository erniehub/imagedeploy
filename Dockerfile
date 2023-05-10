ARG HELM_INSTALL_IMAGE_VERSION

FROM "registry.gitlab.com/gitlab-org/cluster-integration/helm-install-image:${HELM_INSTALL_IMAGE_VERSION}"

# https://github.com/sgerrand/alpine-pkg-glibc
ARG GLIBC_VERSION

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

# Install legacy glibc dependency on amd64
RUN \
  if [[ "$TARGETARCH" = "amd64" ]]; then \
    curl -sSL -o /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub \
      && curl -sSL -O https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VERSION}/glibc-${GLIBC_VERSION}.apk \
      && apk add glibc-${GLIBC_VERSION}.apk \
      && rm glibc-${GLIBC_VERSION}.apk; \
  fi

COPY src/ build/
COPY assets/ assets/

RUN ln -s /build/bin/* /usr/local/bin/
