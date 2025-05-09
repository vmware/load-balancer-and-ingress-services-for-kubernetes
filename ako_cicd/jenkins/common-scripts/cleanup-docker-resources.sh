#!/bin/bash -xe

echo "cleanup script"

# Force remove all containers - both running and stopped
sudo docker ps -a -q | xargs -r sudo docker rm -f

# Remove all unused images, containers, and volumes
sudo docker system prune -a -f
