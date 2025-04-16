#!/bin/bash

set -xe

export GOLANG_SRC_REPO=${PVT_DOCKER_REGISTRY}/dockerhub-proxy-cache/library/golang:latest
export PHOTON_SRC_REPO=${PVT_DOCKER_REGISTRY}/dockerhub-proxy-cache/library/photon:5.0

branch_version=$($WORKSPACE/hack/jenkins/get_branch_version.sh)
version_numbers=(${branch_version//./ })
minor_version=${version_numbers[1]}

make build
make BUILD_TAG=$version_tag docker
# Commenting as we do not require to create operator docker image.
# make BUILD_TAG=$version_tag ako-operator-docker

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
