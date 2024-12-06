#!/bin/bash

set -xe

CI_REGISTRY_PATH=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY
BRANCH=$branch

echo $(git rev-parse origin/${branch}) > $WORKSPACE/HEAD_COMMIT;
cat $WORKSPACE/HEAD_COMMIT

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

#Save ako images as tarball
branch_version=$($WORKSPACE/hack/jenkins/get_branch_version.sh)
version_numbers=(${branch_version//./ })
minor_version=${version_numbers[1]}
sudo docker save -o ako.tar ako:latest
sudo cp -r ako.tar $target_path/
sudo chmod 744 $target_path/ako.tar
if [ "$minor_version" -ge "11" ]; then
	sudo docker save -o ako-operator.tar ako-operator:latest
	sudo docker save -o ako-gateway-api.tar ako-gateway-api:latest
	sudo cp -r ako-operator.tar $target_path/
	sudo cp -r ako-gateway-api.tar $target_path/
	sudo chmod 744 $target_path/ako-operator.tar $target_path/ako-gateway-api.tar
fi

echo "Docker image tar files generated and stored succssfully..."
