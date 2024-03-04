#!/bin/bash

set -xe

export GOLANG_SRC_REPO=${PVT_DOCKER_REGISTRY}/golang:latest
export PHOTON_SRC_REPO=${VMWARE_DOCKER_REGISTRY}/photon/photon4:latest

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
