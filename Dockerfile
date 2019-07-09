FROM alpine:latest

ARG HELM_VERSION
ARG KUBERNETES_VERSION

COPY src/ build/

RUN /build/install_dependencies.sh
RUN ln -s /build/bin/check_kube_domain /usr/local/bin/
