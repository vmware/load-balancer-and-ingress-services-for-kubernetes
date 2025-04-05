#!/bin/bash

set -xe

SRP_SCRIPT_DIR=$WORKSPACE/ako_cicd/jenkins/ci-build/srp
SRP_WORKING_DIR=$WORKSPACE/provenance
if [ "$SRP_UPDATE" = true ]; then
    [ -d "$SRP_WORKING_DIR" ] && sudo rm -rf "$SRP_WORKING_DIR"
    mkdir -p $SRP_WORKING_DIR
    sh ${SRP_SCRIPT_DIR}/srp_cleanup.sh
    sh ${SRP_SCRIPT_DIR}/srp_install.sh
    sudo /srp-tools/observer/bin/observer_agent -m start_observer --output_environment ${SRP_WORKING_DIR}/envs.sh --env_to_shell
    source ${SRP_WORKING_DIR}/envs.sh
fi

echo "--- Start of Pre-Build Steps ---"

# Setting GO related variables for Broadcom GOPROXY artifactory
go env -w GOPROXY=https://packages.vcfd.broadcom.net/artifactory/proxy-golang-remote
go env -w GOSUMDB=on

sudo go clean -modcache

echo "--- End of Pre-Build Steps ---"

echo "--- Start of Build Steps ---"

go mod download

# Setting GO related variables to default values
go env -w GOPROXY=https://proxy.golang.org,direct
go env -w GOSUMDB=sum.golang.org
echo "--- End of Build Steps ---"

cd $WORKSPACE
if [ "$SRP_UPDATE" = true ]; then
    #stop observer and collect network provenance data
    sudo /srp-tools/observer/bin/observer_agent -m stop_observer -f ${SRP_WORKING_DIR}/network_provenance.json
    # Unset the environment variables and cleanup
    source ${SRP_WORKING_DIR}/envs.sh unset
    rm -f ${SRP_WORKING_DIR}/envs.sh
fi
