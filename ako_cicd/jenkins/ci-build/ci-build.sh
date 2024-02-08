#!/bin/bash

set -xe

export GOLANG_SRC_REPO=${PVT_DOCKER_REGISTRY}/golang:latest
#export PHOTON_SRC_REPO=${PVT_DOCKER_REGISTRY}/photon:latest
export PHOTON_SRC_REPO=projects.registry.vmware.com/photon/photon4:latest

make build

echo "--- Start of SRP Cli Tool Installation Steps ---"
sudo rm -rf /srp-tools
sudo mkdir /srp-tools
#sudo wget --config=~/.wgetrc --tries=5 --timeout=60 --quiet --output-document /srp-tools/srp  https://artifactory.eng.vmware.com/artifactory/helix-docker-local/cli/srpcli/0.2.20220623164601-b8f6c65-20/linux/srp
sudo wget --tries=5 --timeout=60 --quiet --output-document /srp-tools/srp https://artifactory.eng.vmware.com/artifactory/srp-tools-generic-local/srpcli/0.5.17-20230303170745-2a66bc5-93/linux-amd64/srp
sudo chmod +x /srp-tools/srp
sudo /srp-tools/srp --version
sudo /srp-tools/srp update --yes
sudo /srp-tools/srp --version

#sudo wget --config=~/.wgetrc --tries=5 --timeout=60 --quiet --output-document /tmp/linux-observer-1.0.4.tar.gz  https://artifactory.eng.vmware.com/osspicli-local/observer/linux-observer-1.0.4.tar.gz
sudo wget --tries=5 --timeout=60 --quiet --output-document /tmp/linux-observer-2.0.0.tar.gz https://artifactory.eng.vmware.com/artifactory/srp-tools-generic-local/observer/2.0.0/linux-observer-2.0.0.tar.gz
sudo mkdir /srp-tools/observer
cd /srp-tools/observer
#sudo tar zxf /tmp/linux-observer-1.0.4.tar.gz
sudo tar zxf /tmp/linux-observer-2.0.0.tar.gz
ls /srp-tools/observer/bin/observer_agent.bash
sudo /srp-tools/observer/bin/observer_agent --version

[ -d "$WORKSPACE/provenance" ] && sudo rm -rf "$WORKSPACE/provenance"
sudo mkdir $WORKSPACE/provenance

echo "--- End of SRP Cli Tool Installation Steps ---"

echo "--- Start of Pre-Build Steps ---"

# Setting GO related variables for VMware's GOPROXY artifactory
go env -w GOPROXY=build-artifactory.eng.vmware.com/artifactory/srp-mds-go-remote
go env -w GOSUMDB=off

sudo go clean -modcache

echo "--- End of Pre-Build Steps ---"

echo "--- Start of Build Steps ---"

cd $WORKSPACE
if [ -z $(sudo lsof -t -i:8989) ]
then
    echo "no mitmproxy process to kill"
else
    sudo kill -9 $(sudo lsof -t -i:8989)
    echo "mitmproxy process killed"
fi

sudo /srp-tools/observer/bin/observer_agent -m start_observer --output_environment $WORKSPACE/provenance/envs.sh --env_to_shell
source $WORKSPACE/provenance/envs.sh
#sudo /srp-tools/observer/bin/observer_agent.bash -t -o $WORKSPACE/provenance/ -- go mod download -x
go mod download
sudo /srp-tools/observer/bin/observer_agent -m stop_observer -f $WORKSPACE/provenance/network_provenance.json
source $WORKSPACE/provenance/envs.sh unset

# Setting GO related variables to default values
go env -w GOPROXY=https://proxy.golang.org,direct
go env -w GOSUMDB=sum.golang.org
echo "--- End of Build Steps ---"

if [ "$RUN_TESTS" = true ]; then
    make test
fi


if [ "$RUN_INT_TESTS" = true ]; then
    make int_test
fi
