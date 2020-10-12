#!/bin/sh

# Copyright (c) 2020, Oracle and/or its affiliates.
#
# Licensed under the Universal Permissive License v 1.0 as shown at https://oss.oracle.com/licenses/upl/

set -o errexit
set -o nounset

if [ -z "${PKG}" ]; then
    echo "PKG must be set"
    exit 1
fi
if [ -z "${ARCH}" ]; then
    echo "ARCH must be set"
    exit 1
fi
if [ -z "${OS}" ]; then
    echo "OS must be set"
    exit 1
fi
if [ -z "${VERSION}" ]; then
    echo "VERSION must be set"
    exit 1
fi

export CGO_ENABLED=0
export GOARCH="${ARCH}"
export GOOS="${OS}"

go install                                                   \
   -v                                                        \
    -installsuffix "static"                                  \
    -ldflags "-X ${PKG}/pkg/version.buildVersion=${VERSION}" \
    ./...
