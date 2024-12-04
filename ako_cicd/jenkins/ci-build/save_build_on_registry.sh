#!/bin/bash

set -xe

sudo docker images

version_tag=$($WORKSPACE/hack/jenkins/get_build_version.sh $JOB_NAME $BUILD_NUMBER)
branch_version=$($WORKSPACE/hack/jenkins/get_branch_version.sh)
version_numbers=(${branch_version//./ })
minor_version=${version_numbers[1]}

AKO_IMAGES=($DOCKER_AKO_IMAGE_NAME $DOCKER_AKO_OPERATOR_IMAGE_NAME)

if [ "$minor_version" -ge "11" ]; then
    AKO_IMAGES+=($DOCKER_AKO_GATEWAY_API_IMAGE_NAME)
fi

echo ${AKO_IMAGES[@]}
branch=av-223899_master
for image in "${AKO_IMAGES[@]}"
do
  source_image=$image:latest
  target_image=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY/ako/$branch/$image:$version_tag
  sudo docker tag $source_image $target_image
  sudo docker push $target_image
done
