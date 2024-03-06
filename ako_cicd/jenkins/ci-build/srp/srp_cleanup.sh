#!/bin/bash
set -xe

#stop observer 
SRP_WORKING_DIR=$WORKSPACE/provenance
#stop observer and collect network provenance data
echo "stopping observer, generating network_provenance.json and cleaning up"
sudo /srp-tools/observer/bin/observer_agent -m stop_observer -f ${SRP_WORKING_DIR}/network_provenance.json || true
# Unset the environment variables and cleanup
source ${SRP_WORKING_DIR}/envs.sh unset || true
rm -f ${SRP_WORKING_DIR}/envs.sh || true
sudo rm -rf /srp-tools
