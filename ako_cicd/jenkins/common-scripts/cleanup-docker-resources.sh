#!/bin/bash -xe

# Force remove all containers - both running and stopped
sudo docker ps -a -q | xargs -r sudo docker rm -f

# Remove all unused images, containers, and volumes
sudo docker system prune -a -f

#Remove all the build cache data
sudo docker builder prune -a -f

#pring docker system usage on disk
sudo docker system df
