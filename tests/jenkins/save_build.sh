#!/bin/bash

set -xe


if [ $# -lt 2 ] ; then
    echo "Usage: ./save_build.sh <BRANCH> <BUILD_NUMBER>";
    exit 1
fi

BRANCH=$1
BUILD_NUMBER=$2


SCRIPTPATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# Function to get GIT workspace root location
function get_git_ws {
    git_ws=$(git rev-parse --show-toplevel)
    [ -z "$git_ws" ] && echo "Couldn't find git workspace root" && exit 1
    echo $git_ws
}

BUILD_VERSION_SCRIPT=$SCRIPTPATH/get_build_version.sh
CHARTS_PATH="$(get_git_ws)/helm/ako"

build_version=$(bash $BUILD_VERSION_SCRIPT "dummy" $BUILD_NUMBER)

target_path=/mnt/builds/ako/$BRANCH/ci-build-$build_version

sudo mkdir -p $target_path

sudo cp -r $CHARTS_PATH/* $target_path/

set +e
sudo cp "$(get_git_ws)/HEAD_COMMIT" $target_path/

if [ "$?" != "0" ]; then
    echo "ERROR: Could not save the head commit file into target path"
fi

set -e

sudo sed -i --regexp-extended "s/^(\s*)(appVersion\s*:\s*latest\s*$)/\1appVersion: $build_version/" $target_path/Chart.yaml
