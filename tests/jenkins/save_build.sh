#!/bin/bash

set -xe


if [ $# -lt 1 ] ; then
    echo "Usage: ./save_build.sh <BUILD_NUMBER>";
    exit 1
fi

BUILD_NUMBER=$1


SCRIPTPATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# Function to get GIT workspace root location
function get_git_ws {
    git_ws=$(git rev-parse --show-toplevel)
    [ -z "$git_ws" ] && echo "Couldn't find git workspace root" && exit 1
    echo $git_ws
}


BRANCH_VERSION_SCRIPT=$SCRIPTPATH/get_branch_version.sh
BUILD_VERSION_SCRIPT=$SCRIPTPATH/get_build_version.sh
CHARTS_PATH="$(get_git_ws)/helm/ako"

build_version=$(bash $BUILD_VERSION_SCRIPT "dummy" $BUILD_NUMBER)
branch_version=$(bash $BRANCH_VERSION_SCRIPT)

target_path=/mnt/builds/ako/$branch_version/nightly-build-$build_version

sudo mkdir -p $target_path

sudo cp -r $CHARTS_PATH/* $target_path/

sudo sed -i --regexp-extended "s/^(\s*)(version\s*:\s*0.1.0\s*$)/\1version: $build_version/" $target_path/Chart.yaml
