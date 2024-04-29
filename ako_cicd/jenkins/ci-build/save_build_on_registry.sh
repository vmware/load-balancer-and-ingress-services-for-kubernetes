#!/bin/bash

set -xe

version_tag=$($WORKSPACE/hack/jenkins/get_build_version.sh $JOB_NAME $BUILD_NUMBER)

sudo docker images

source_image=$DOCKER_AKO_IMAGE_NAME:latest

target_image=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY/$DOCKER_AKO_IMAGE_NAME:$version_tag

sudo docker tag $source_image $target_image

sudo docker push $target_image


source_image=$DOCKER_AKO_OPERATOR_IMAGE_NAME:latest

target_image=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY/$DOCKER_AKO_OPERATOR_IMAGE_NAME:$version_tag

sudo docker tag $source_image $target_image

sudo docker push $target_image


source_image=$DOCKER_AKO_GATEWAY_API_IMAGE_NAME:latest

target_image=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY/$DOCKER_AKO_GATEWAY_API_IMAGE_NAME:$version_tag

sudo docker tag $source_image $target_image

sudo docker push $target_image

