
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


if [ $# -lt 5 ] ; then
    echo "Usage: ./save_build.sh <BRANCH> <BUILD_NUMBER> <WORKSPACE> <JENKINS_JOB_NAME> <JENKINS_URL>";
    exit 1
fi

BRANCH=$1
BUILD_NUMBER=$2
WORKSPACE=$3
JENKINS_JOB_NAME=$4
JENKINS_URL=$5

SCRIPTPATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"

# Function to get GIT workspace root location
function get_git_ws {
    git_ws=$(git rev-parse --show-toplevel)
    [ -z "$git_ws" ] && echo "Couldn't find git workspace root" && exit 1
    echo $git_ws
}

BRANCH_VERSION_SCRIPT=$SCRIPTPATH/get_branch_version.sh
# Compute base_build_num
base_build_num=$(cat $(get_git_ws)/base_build_num)
version_build_num=$(expr "$base_build_num" + "$BUILD_NUMBER")
branch_version=$(bash $BRANCH_VERSION_SCRIPT)

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

#collecting source provenance data
PRODUCT_NAME="Avi Kubernetes Operator"
JENKINS_INSTANCE=$(echo $JENKINS_URL | sed -E 's/^\s*.*:\/\///g' | sed -E 's/:.*//g')
COMP_UID="uid.obj.build.jenkins(instance='$JENKINS_INSTANCE',job_name='$JENKINS_JOB_NAME',build_number='$BUILD_NUMBER')"
provenance_source_file="$WORKSPACE/provenance/source.json"

# Function to run srp source provenance command
function source_provenance() {
    sudo /srp-tools/srp provenance source --scm-type git --name "$PRODUCT_NAME" --path ./ --saveto $provenance_source_file --comp-uid $COMP_UID --build-number ${version_build_num} --version $branch_version --all-ephemeral true --build-type release $@
}

output=( $(find $WORKSPACE/ -type d  -not -path "$WORKSPACE/build/*" -name '.git') )
for line in "${output[@]}"
do
    cd $(dirname $line)
    if [ -f  $provenance_source_file ]
    then
        source_provenance --append
    else
        source_provenance
    fi
done
cd $WORKSPACE
sudo /srp-tools/srp provenance merge --source ./provenance/source.json --network ./provenance/provenance.json --saveto ./provenance/merged.json
provenance_path=$target_path/provenance
sudo mkdir -p $provenance_path
sudo cp $WORKSPACE/provenance/*json $provenance_path/;
