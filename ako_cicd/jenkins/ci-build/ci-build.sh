#!/bin/bash

set -xe

export GOLANG_SRC_REPO=avi-alb-docker-virtual.packages.vcfd.broadcom.net/golang:latest
export PHOTON_SRC_REPO=photonos-docker-local.packages.vcfd.broadcom.net/photon5:latest

# branch_version=$($WORKSPACE/hack/jenkins/get_branch_version.sh)
# version_numbers=(${branch_version//./ })
# minor_version=${version_numbers[1]}

export PATH=$PATH:/usr/local/go/bin
go version

make build
make BUILD_TAG=$version_tag docker
make BUILD_TAG=$version_tag ako-operator-docker

make BUILD_TAG=$version_tag build-gateway-api
make BUILD_TAG=$version_tag ako-gateway-api-docker


if [ "$RUN_TESTS" = true ]; then
    make test
fi

if [ "$RUN_INT_TESTS" = true ]; then
    make int_test
fi

#Steps to Build and Test AKO-CRD-OPERATOR
cd $WORKSPACE/ako-crd-operator
go env -w GOSUMDB=off
make lint
make build
make BUILD_TAG=$version_tag docker-build
#make test
