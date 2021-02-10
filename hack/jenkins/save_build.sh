
# Copyright 2019-2020 VMware, Inc.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

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
