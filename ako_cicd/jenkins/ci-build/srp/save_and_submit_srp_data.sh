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

BRANCH=$branch
CI_REGISTRY_PATH=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY

BUILD_VERSION_SCRIPT=$WORKSPACE/hack/jenkins/get_build_version.sh
build_version=$(bash $BUILD_VERSION_SCRIPT "dummy" $BUILD_NUMBER)

target_path=/mnt/builds/ako_OS/$BRANCH/ci-build-$build_version

#collecting provenance data
PRODUCT_NAME="Avi Kubernetes Operator"
JENKINS_INSTANCE=$(echo $JENKINS_URL | sed -E 's/^\s*.*:\/\///g' | sed -E 's/:.*//g')
COMP_UID="uid.obj.build.jenkins(instance='$JENKINS_INSTANCE',job_name='$JENKINS_JOB_NAME',build_number='$BUILD_NUMBER')"

# initialize credentials that are required for submission, Credentials value set by jenkins vault plugin
sudo /srp-tools/srp config auth --client-id=${SRP_CLIENT_ID} --client-secret=${SRP_CLIENT_SECRECT}

# initialize blank provenance in the working directory, $SRP_WORKING_DIR
sudo /srp-tools/srp provenance init --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-build jenkins --instance $JENKINS_INSTANCE --build-number $BUILD_NUMBER --job-name $JENKINS_JOB_NAME --working-dir $WORKSPACE/provenance

# add an action for the golang build, importing the observations that were captured in the build-golang-app step
sudo /srp-tools/srp provenance action start --name=ako-build --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance action import-observation --name=ako-obs --file=$WORKSPACE/provenance/network_provenance.json --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance action stop --working-dir $WORKSPACE/provenance

# declare the git source tree for the build.  We refer to this declaration below when adding source inputs.
sudo /srp-tools/srp provenance declare-source git --verbose --set-key=mainsrc --path=$WORKSPACE --branch=$BRANCH --working-dir $WORKSPACE/provenance

CI_REGISTRY_IMAGE_AKO=$CI_REGISTRY_PATH/ako
IMAGE_DIGEST=`sudo docker images $CI_REGISTRY_IMAGE_AKO  --digests | grep sha256 | xargs | cut -d " " -f3`
echo $IMAGE_DIGEST
sudo /srp-tools/srp provenance add-output package.oci --set-key=ako-image --action-key=ako-build --name=${CI_REGISTRY_IMAGE_AKO}  --digest=${IMAGE_DIGEST} --working-dir $WORKSPACE/provenance

CI_REGISTRY_IMAGE_AKO_OPERATOR=$CI_REGISTRY_PATH/ako-operator
IMAGE_DIGEST=`sudo docker images $CI_REGISTRY_IMAGE_AKO_OPERATOR  --digests | grep sha256 | xargs | cut -d " " -f3`
echo $IMAGE_DIGEST
sudo /srp-tools/srp provenance add-output package.oci --set-key=ako-operator-image --action-key=ako-build --name=${CI_REGISTRY_IMAGE_AKO_OPERATOR}  --digest=${IMAGE_DIGEST} --working-dir $WORKSPACE/provenance

CI_REGISTRY_IMAGE_AKO_GATEWAY_API=$CI_REGISTRY_PATH/ako-gateway-api
IMAGE_DIGEST=`sudo docker images $CI_REGISTRY_IMAGE_AKO_GATEWAY_API  --digests | grep sha256 | xargs | cut -d " " -f3`
echo $IMAGE_DIGEST
sudo /srp-tools/srp provenance add-output package.oci --set-key=ako-gateway-api-image --action-key=ako-build --name=${CI_REGISTRY_IMAGE_AKO_GATEWAY_API}  --digest=${IMAGE_DIGEST} --working-dir $WORKSPACE/provenance

# use the syft plugin to scan the container and add all inputs it discovers. This will include the golang application we added
# to the container, which are duplicate of the inputs above, but in this case we KNOW they are incorporated.
sudo /srp-tools/srp provenance add-input syft --output-key=ako-image --usage functionality --incorporated true --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-input syft --output-key=ako-operator-image --usage functionality --incorporated true --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-input syft --output-key=ako-gateway-api-image --usage functionality --incorporated true --working-dir $WORKSPACE/provenance

# adding source input
sudo /srp-tools/srp provenance add-input source --source-key=mainsrc --output-key=ako-image --is-component-source --incorporated=true --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-input source --source-key=mainsrc --output-key=ako-operator-image --is-component-source --incorporated=true --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-input source --source-key=mainsrc --output-key=ako-gateway-api-image --is-component-source --incorporated=true --working-dir $WORKSPACE/provenance

# compile the provenance to a file and then dump it out to the console for reference
sudo /srp-tools/srp provenance compile --saveto $WORKSPACE/provenance/srp_prov3_fragment.json --working-dir $WORKSPACE/provenance
cat $WORKSPACE/provenance/srp_prov3_fragment.json

# submit the created provenance to SRP
sudo /srp-tools/srp provenance submit --verbose --path $WORKSPACE/provenance/srp_prov3_fragment.json --working-dir $WORKSPACE/provenance

provenance_path=$target_path/provenance
sudo mkdir -p $provenance_path
sudo cp $WORKSPACE/provenance/* $provenance_path/;
