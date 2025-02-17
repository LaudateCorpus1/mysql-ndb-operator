# Copyright (c) 2021, Oracle and/or its affiliates.
#
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

FROM oraclelinux:8-slim

# install go
RUN echo "[main]" > /etc/dnf/dnf.conf && \
    microdnf install go-toolset

# install kubectl
ENV KUBECTL_BASE_URL https://dl.k8s.io/release
ENV KUBECTL_BINARY bin/linux/amd64/kubectl
RUN curl -LO \
 "$KUBECTL_BASE_URL/$(curl -L -s $KUBECTL_BASE_URL/stable.txt)/$KUBECTL_BINARY"
RUN install -o root -g root -m 0755 kubectl /usr/local/bin/kubectl

WORKDIR /ndb-operator-e2e-testing/
# copy all files/directories in project directory to workdir
# excludes files from Dockerfile.dockerignore
COPY . .

ENTRYPOINT ["go", "run", "github.com/onsi/ginkgo/ginkgo", "-r", "-keepGoing", "e2e-tests/suites"]
