#!/bin/bash
set -xe

#SRP intallation
sudo rm -rf /srp-tools
sudo mkdir /srp-tools
sudo wget --tries=5 --timeout=60 --quiet --output-document /srp-tools/srp https://packages.vcfd.broadcom.net/artifactory/srp-generic-local/srpcli/1.34.3-20250923143853-9177b56/linux-amd64/srp
sudo chmod +x /srp-tools/srp
sudo /srp-tools/srp --version
sudo /srp-tools/srp update --yes
sudo /srp-tools/srp --version
sudo wget --tries=5 --timeout=60 --quiet --output-document /tmp/linux-observer-4.4.0.tar.gz https://packages.vcfd.broadcom.net/artifactory/srp-generic-local/observer/4.4.0/linux-observer-4.4.0.tar.gz
sudo mkdir /srp-tools/observer
cd /srp-tools/observer
sudo tar zxf /tmp/linux-observer-4.4.0.tar.gz
sudo /srp-tools/observer/bin/observer_agent --version
