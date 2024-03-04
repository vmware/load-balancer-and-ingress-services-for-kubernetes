#!/bin/bash

set -xe

cd $WORKSPACE/hack/jenkins;
echo $GIT_COMMIT > $WORKSPACE/HEAD_COMMIT;

CI_REGISTRY_PATH=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY
BRANCH=$branch

# Function to get GIT workspace root location
function get_git_ws {
    git_ws=$(git rev-parse --show-toplevel)
    [ -z "$git_ws" ] && echo "Couldn't find git workspace root" && exit 1
    echo $git_ws
}


BUILD_VERSION_SCRIPT=$WORKSPACE/hack/jenkins/get_build_version.sh
CHARTS_PATH="$(get_git_ws)/helm/ako"
AKO_OPERATOR_CHARTS_PATH="$(get_git_ws)/ako-operator/helm/ako-operator"

build_version=$(bash $BUILD_VERSION_SCRIPT "dummy" $BUILD_NUMBER)

target_path=/mnt/builds/ako_OS/$BRANCH/ci-build-$build_version
ako_operator_target_path=$target_path/ako-operator

sudo mkdir -p $target_path
sudo mkdir -p $ako_operator_target_path

sudo cp -r $CHARTS_PATH/* $target_path/
sudo cp -r $AKO_OPERATOR_CHARTS_PATH/* $ako_operator_target_path/


set +e
sudo cp "$(get_git_ws)/HEAD_COMMIT" $target_path/

if [ "$?" != "0" ]; then
    echo "ERROR: Could not save the head commit file into target path"
fi

set -e

sudo sed -i --regexp-extended "s/^(\s*)(appVersion\s*:\s*latest\s*$)/\1appVersion: $build_version/" $target_path/Chart.yaml
