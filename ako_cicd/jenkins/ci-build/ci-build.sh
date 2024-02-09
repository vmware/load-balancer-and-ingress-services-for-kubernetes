#!/bin/bash

set -xe

SRP_SCRIPT_DIR=$WORKSPACE/ako_cicd/jenkins/ci-build/srp
if [ "$SRP_UPDATE" = true ]; then
    sh ${SRP_SCRIPT_DIR}/srp_install_and_start.sh
fi

export GOLANG_SRC_REPO=${PVT_DOCKER_REGISTRY}/golang:latest
#export PHOTON_SRC_REPO=${PVT_DOCKER_REGISTRY}/photon:latest
export PHOTON_SRC_REPO=projects.registry.vmware.com/photon/photon4:latest

make build
make BUILD_TAG=$version_tag docker
make BUILD_TAG=$version_tag ako-operator-docker
make BUILD_TAG=$version_tag build-gateway-api
make BUILD_TAG=$version_tag ako-gateway-api-docker

echo "--- Start of Pre-Build Steps ---"

# Setting GO related variables for VMware's GOPROXY artifactory
go env -w GOPROXY=build-artifactory.eng.vmware.com/artifactory/srp-mds-go-remote
go env -w GOSUMDB=off

sudo go clean -modcache

echo "--- End of Pre-Build Steps ---"

echo "--- Start of Build Steps ---"

cd $WORKSPACE
if [ -z $(sudo lsof -t -i:8989) ]
then
    echo "no mitmproxy process to kill"
else
    sudo kill -9 $(sudo lsof -t -i:8989)
    echo "mitmproxy process killed"
fi

go mod download

# Setting GO related variables to default values
go env -w GOPROXY=https://proxy.golang.org,direct
go env -w GOSUMDB=sum.golang.org
echo "--- End of Build Steps ---"

cd $WORKSPACE
if [ "$SRP_UPDATE" = true ]; then
    #stop observer and collect network provenance data
    sh ${SRP_SCRIPT_DIR}/srp_stop_and_cleanup.sh
fi

if [ "$RUN_TESTS" = true ]; then
    make test
fi

if [ "$RUN_INT_TESTS" = true ]; then
    make int_test
fi
