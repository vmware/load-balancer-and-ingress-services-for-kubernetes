# Copyright © 2025 Broadcom Inc. and/or its subsidiaries. All Rights Reserved.
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

set -e


if [ $# -lt 2 ] ; then
    echo "Usage: ./get_build_version.sh <JOB> <BUILD_NUMBER>";
    exit 1
fi

JOB=$1
BUILD_NUMBER=$2


SCRIPTPATH="$( cd "$(dirname "$0")" >/dev/null 2>&1 ; pwd -P )"


BRANCH_VERSION_SCRIPT=$SCRIPTPATH/get_branch_version.sh

# Function to get GIT workspace root location
function get_git_ws {
    git_ws=$(git rev-parse --show-toplevel)
    [ -z "$git_ws" ] && echo "Couldn't find git workspace root" && exit 1
    echo $git_ws
}

# Pull the major, minor, maintenance versions from the repository's version.yaml file
version_file=$(get_git_ws)/version.yaml

# Compute base_build_num
base_build_num=$(cat $(get_git_ws)/base_build_num)
version_build_num=$(expr "$base_build_num" + "$BUILD_NUMBER")
branch_version=$(bash $BRANCH_VERSION_SCRIPT)
version_tag="$branch_version-$version_build_num"

mkdir -p /tmp/$JOB;
touch /tmp/$JOB/jenkins.properties;
echo "version_tag=${version_tag}" > /tmp/$JOB/jenkins.properties;
echo $version_tag
