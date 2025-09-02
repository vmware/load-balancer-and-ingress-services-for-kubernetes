#!/bin/bash

set -xe

IFS=', ' read -r -a registries <<< "$REMOTE_DOCKER_REGISTRIES"

if [ "${#registries[@]}" == "0" ]; then
	echo "Remote registries count must be non-zero"
    exit 1
fi

branch_version=$($WORKSPACE/hack/jenkins/get_branch_version.sh)
version_tag=$($WORKSPACE/hack/jenkins/get_build_version.sh $JOB_NAME $build_num)



VENV_PATH=$HOME/venv
if [ ! -d "$VENV_PATH" ]; then
	virtualenv -p python3 $VENV_PATH
fi

source $VENV_PATH/bin/activate
pip install sshuttle==1.0.4

set +e
pgrep -f sshuttle | xargs -r sudo kill -9



ps -ef | grep sshuttle | grep -v grep
rc=$?

set -e

if [ "$rc" != "0" ]; then
    sshuttle -D -r $JUMPHOST_USER@$JUMPHOST_IP $JUMPHOST_PROXY/$JUMPHOST_PROXY_PORT -e "ssh -i  $SSH_PVT_KEY_FILE" -v
fi

###########
AKO_IMAGES=("ako" "ako-operator" "ako-crd-operator")
version_numbers=(${branch_version//./ })
minor_version=${version_numbers[1]}

if [ "$minor_version" -ge "11" ]; then
    AKO_IMAGES+=("ako-gateway-api")
fi

echo ${AKO_IMAGES[@]}

for image in "${AKO_IMAGES[@]}"
do
    source_image=$PVT_DOCKER_REGISTRY/$PVT_DOCKER_REPOSITORY/ako/$branch/$image:$version_tag
    sudo docker pull $source_image
    for registry in "${registries[@]}"
    do
        target_image="$registry/$PVT_DOCKER_REPOSITORY/ako/$branch/$image:$version_tag"
        echo "Tagging and pushing to registry: $registry"
        sudo docker tag $source_image $target_image
        sudo docker push $target_image
    done
done
