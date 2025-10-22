#!/bin/bash

set -xe

export GOLANG_SRC_REPO=avi-alb-docker-virtual.packages.vcfd.broadcom.net/golang:latest
export PHOTON_SRC_REPO=photonos-docker-local.packages.vcfd.broadcom.net/photon5-amd64:20251011

branch_version=$($WORKSPACE/hack/jenkins/get_branch_version.sh)
version_numbers=(${branch_version//./ })
minor_version=${version_numbers[1]}

export PATH=$PATH:/usr/local/go/bin
go version

make build
make BUILD_TAG=$version_tag docker
make BUILD_TAG=$version_tag ako-operator-docker

if [ "$minor_version" -ge "11" ]; then
    make BUILD_TAG=$version_tag build-gateway-api
    make BUILD_TAG=$version_tag ako-gateway-api-docker
fi

if [ "$RUN_TESTS" = true ]; then
    make test
fi

if [ "$RUN_INT_TESTS" = true ]; then
    make int_test
fi
