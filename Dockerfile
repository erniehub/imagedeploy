ARG HELM_VERSION
ARG KUBERNETES_VERSION

#FROM "registry.gitlab.com/gitlab-org/cluster-integration/helm-install-image/releases/${HELM_VERSION}-kube-${KUBERNETES_VERSION}"
FROM "registry.gitlab.com/gitlab-org/cluster-integration/helm-install-image/branches/kubectl-1-16-7:d217f17e305cd2cfada4073aa33fbf3aa48bffc9-helm2"
## Remove hard-coded FROM once the helm-install-image is tagged for release. 

# https://github.com/sgerrand/alpine-pkg-glibc
ARG GLIBC_VERSION

# Install Dependencies
RUN apk add --no-cache openssl curl tar gzip bash jq \
  && curl -sSL -o /etc/apk/keys/sgerrand.rsa.pub https://alpine-pkgs.sgerrand.com/sgerrand.rsa.pub \
  && curl -sSL -O https://github.com/sgerrand/alpine-pkg-glibc/releases/download/${GLIBC_VERSION}/glibc-${GLIBC_VERSION}.apk \
  && apk add glibc-${GLIBC_VERSION}.apk \
  && apk add ruby jq ruby-json \
  && rm glibc-${GLIBC_VERSION}.apk

COPY src/ build/
COPY assets/ assets/

RUN ln -s /build/bin/* /usr/local/bin/
