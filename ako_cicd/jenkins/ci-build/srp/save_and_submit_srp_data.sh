#!/bin/bash

set -xe

BRANCH=$branch
CI_REGISTRY_PATH=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY

BUILD_VERSION_SCRIPT=$WORKSPACE/hack/jenkins/get_build_version.sh
build_version=$(bash $BUILD_VERSION_SCRIPT "dummy" $BUILD_NUMBER)

target_path=/mnt/builds/ako_OS/$BRANCH/ci-build-$build_version

sudo mkdir -p $target_path

#collecting provenance data
PRODUCT_NAME="Avi Kubernetes Operator"
JENKINS_INSTANCE=$(echo $JENKINS_URL | sed -E 's/^\s*.*:\/\///g' | sed -E 's/\///g')
COMP_UID="uid.obj.build.jenkins(instance='$JENKINS_INSTANCE',job_name='$JOB_NAME',build_number='$BUILD_NUMBER')"

# initialize credentials that are required for submission, Credentials value set by jenkins vault plugin
sudo /srp-tools/srp config auth --client-id=${SRP_CLIENT_ID} --client-secret=${SRP_CLIENT_SECRECT}

# initialize blank provenance in the working directory, $SRP_WORKING_DIR
sudo /srp-tools/srp provenance init --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-build jenkins --instance $JENKINS_INSTANCE --build-number $BUILD_NUMBER --job-name $JOB_NAME --working-dir $WORKSPACE/provenance

# add an action for the golang build, importing the observations that were captured in the build-golang-app step
sudo /srp-tools/srp provenance action start --name=ako-build --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance action import-observation --name=ako-obs --file=$WORKSPACE/provenance/network_provenance.json --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance action stop --working-dir $WORKSPACE/provenance

# declare the git source tree for the build.  We refer to this declaration below when adding source inputs.
sudo /srp-tools/srp provenance declare-source git --verbose --set-key=mainsrc --path=$WORKSPACE --branch=$BRANCH --working-dir $WORKSPACE/provenance

#Enable this option to create image manifest.json
export DOCKER_CLI_EXPERIMENTAL=enabled

CI_REGISTRY_IMAGE_AKO=$CI_REGISTRY_PATH/ako/$branch/ako
IMAGE_DIGEST=`sudo docker images $CI_REGISTRY_IMAGE_AKO  --digests | grep sha256 | xargs | cut -d " " -f3`
echo $IMAGE_DIGEST
docker manifest inspect $CI_REGISTRY_IMAGE_AKO:${build_version} --insecure > ako_manifest.json
cat ako_manifest.json
sudo /srp-tools/srp provenance add-output package.oci --set-key=ako-image --action-key=ako-build --name=${CI_REGISTRY_IMAGE_AKO}  --digest=${IMAGE_DIGEST} --manifest-path $WORKSPACE/ako_manifest.json --working-dir $WORKSPACE/provenance

CI_REGISTRY_IMAGE_AKO_OPERATOR=$CI_REGISTRY_PATH/ako/$branch/ako-operator
IMAGE_DIGEST=`sudo docker images $CI_REGISTRY_IMAGE_AKO_OPERATOR  --digests | grep sha256 | xargs | cut -d " " -f3`
echo $IMAGE_DIGEST
docker manifest inspect $CI_REGISTRY_IMAGE_AKO_OPERATOR:${build_version} --insecure > ako_operator_manifest.json
cat ako_operator_manifest.json
sudo /srp-tools/srp provenance add-output package.oci --set-key=ako-operator-image --action-key=ako-build --name=${CI_REGISTRY_IMAGE_AKO_OPERATOR}  --digest=${IMAGE_DIGEST} --manifest-path $WORKSPACE/ako_operator_manifest.json --working-dir $WORKSPACE/provenance

CI_REGISTRY_IMAGE_AKO_CRD_OPERATOR=$CI_REGISTRY_PATH/ako/$branch/ako-crd-operator
IMAGE_DIGEST=`sudo docker images $CI_REGISTRY_IMAGE_AKO_CRD_OPERATOR  --digests | grep sha256 | xargs | cut -d " " -f3`
echo $IMAGE_DIGEST
docker manifest inspect $CI_REGISTRY_IMAGE_AKO__CRD_OPERATOR:${build_version} --insecure > ako_crd_operator_manifest.json
cat ako_crd_operator_manifest.json
sudo /srp-tools/srp provenance add-output package.oci --set-key=ako-crd-operator-image --action-key=ako-build --name=${CI_REGISTRY_IMAGE_AKO_CRD_OPERATOR}  --digest=${IMAGE_DIGEST} --manifest-path $WORKSPACE/ako_crd_operator_manifest.json --working-dir $WORKSPACE/provenance

branch_version=$($WORKSPACE/hack/jenkins/get_branch_version.sh)
version_numbers=(${branch_version//./ })
minor_version=${version_numbers[1]}

if [ "$minor_version" -ge "11" ]; then
    CI_REGISTRY_IMAGE_AKO_GATEWAY_API=$CI_REGISTRY_PATH/ako/$branch/ako-gateway-api
    IMAGE_DIGEST=`sudo docker images $CI_REGISTRY_IMAGE_AKO_GATEWAY_API  --digests | grep sha256 | xargs | cut -d " " -f3`
    echo $IMAGE_DIGEST
    docker manifest inspect $CI_REGISTRY_IMAGE_AKO_GATEWAY_API:${build_version} --insecure > ako_gateway_api_manifest.json
    cat ako_gateway_api_manifest.json
    sudo /srp-tools/srp provenance add-output package.oci --set-key=ako-gateway-api-image --action-key=ako-build --name=${CI_REGISTRY_IMAGE_AKO_GATEWAY_API}  --digest=${IMAGE_DIGEST} --manifest-path $WORKSPACE/ako_gateway_api_manifest.json --working-dir $WORKSPACE/provenance
fi
# use the syft plugin to scan the container and add all inputs it discovers. This will include the golang application we added
# to the container, which are duplicate of the inputs above, but in this case we KNOW they are incorporated.
sudo /srp-tools/srp provenance add-input syft --output-key=ako-image --usage functionality --incorporated true --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-input syft --output-key=ako-operator-image --usage functionality --incorporated true --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-input syft --output-key=ako-crd-operator-image --usage functionality --incorporated true --working-dir $WORKSPACE/provenance
if [ "$minor_version" -ge "11" ]; then
    sudo /srp-tools/srp provenance add-input syft --output-key=ako-gateway-api-image --usage functionality --incorporated true --working-dir $WORKSPACE/provenance
fi

# adding source input
sudo /srp-tools/srp provenance add-input source --source-key=mainsrc --output-key=ako-image --is-component-source --incorporated=true --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-input source --source-key=mainsrc --output-key=ako-operator-image --is-component-source --incorporated=true --working-dir $WORKSPACE/provenance
sudo /srp-tools/srp provenance add-input source --source-key=mainsrc --output-key=ako-crd-operator-image --is-component-source --incorporated=true --working-dir $WORKSPACE/provenance
if [ "$minor_version" -ge "11" ]; then
    sudo /srp-tools/srp provenance add-input source --source-key=mainsrc --output-key=ako-gateway-api-image --is-component-source --incorporated=true --working-dir $WORKSPACE/provenance
fi

# compile the provenance to a file and then dump it out to the console for reference
sudo /srp-tools/srp provenance compile --saveto $WORKSPACE/provenance/srp_prov3_fragment.json --working-dir $WORKSPACE/provenance
cat $WORKSPACE/provenance/srp_prov3_fragment.json

# submit the created provenance to SRP
sudo /srp-tools/srp provenance submit --verbose --path $WORKSPACE/provenance/srp_prov3_fragment.json --working-dir $WORKSPACE/provenance

provenance_path=$target_path/provenance
sudo mkdir -p $provenance_path
sudo cp $WORKSPACE/provenance/* $provenance_path/;
